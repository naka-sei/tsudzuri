package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	entpage "github.com/naka-sei/tsudzuri/infrastructure/db/ent/page"
	entuser "github.com/naka-sei/tsudzuri/infrastructure/db/ent/user"
	"github.com/naka-sei/tsudzuri/pkg/testutil"
)

func TestConnection_RunInTransaction(t *testing.T) {
	type args struct {
		run func(ctx context.Context, conn *Connection) error
	}
	type want struct {
		err    error
		assert func(t *testing.T, ctx context.Context, conn *Connection, data *caseData)
	}

	tests := []struct {
		name    string
		prepare func(*caseData)
		args    func(*caseData) args
		want    func(*caseData) want
	}{
		{
			name: "commit_success",
			prepare: func(d *caseData) {
				d.uid = fmt.Sprintf("txn-user-%s", uuid.NewString())
				d.title = fmt.Sprintf("txn-page-%s", uuid.NewString())
				d.invite = randomInviteCode()
			},
			args: func(d *caseData) args {
				return args{run: func(ctx context.Context, conn *Connection) error {
					return conn.RunInTransaction(ctx, func(txCtx context.Context) error {
						t.Helper()
						if conn.ReadOnlyDB(txCtx) != conn.WriteDB(txCtx) {
							return fmt.Errorf("expected read/write clients to be identical within transaction")
						}

						client := conn.WriteDB(txCtx)
						creator, err := client.User.Create().
							SetUID(d.uid).
							SetProvider(entuser.ProviderAnonymous).
							Save(txCtx)
						if err != nil {
							return err
						}

						_, err = client.Page.Create().
							SetTitle(d.title).
							SetCreatorID(creator.ID).
							SetInviteCode(d.invite).
							Save(txCtx)
						return err
					})
				}}
			},
			want: func(d *caseData) want {
				return want{
					assert: func(t *testing.T, ctx context.Context, conn *Connection, data *caseData) {
						client := conn.ReadOnlyDB(ctx)
						storedUser, err := client.User.Query().Where(entuser.UID(data.uid)).Only(ctx)
						if err != nil {
							t.Fatalf("failed to fetch committed user: %v", err)
						}

						count, err := client.Page.Query().Where(entpage.CreatorIDEQ(storedUser.ID)).Count(ctx)
						if err != nil {
							t.Fatalf("failed to count pages: %v", err)
						}
						if count != 1 {
							t.Fatalf("expected 1 page, got %d", count)
						}
					},
				}
			},
		},
		{
			name: "nested_savepoint_rollback",
			prepare: func(d *caseData) {
				d.outerUID = fmt.Sprintf("outer-user-%s", uuid.NewString())
				d.innerTitle = fmt.Sprintf("inner-page-%s", uuid.NewString())
				d.innerInvite = randomInviteCode()
				d.expectedErr = errors.New("inner failure")
			},
			args: func(d *caseData) args {
				return args{run: func(ctx context.Context, conn *Connection) error {
					return conn.RunInTransaction(ctx, func(txCtx context.Context) error {
						client := conn.WriteDB(txCtx)
						creator, err := client.User.Create().
							SetUID(d.outerUID).
							SetProvider(entuser.ProviderAnonymous).
							Save(txCtx)
						if err != nil {
							return err
						}

						innerErr := conn.RunInTransaction(txCtx, func(innerCtx context.Context) error {
							innerClient := conn.WriteDB(innerCtx)
							if _, err := innerClient.Page.Create().
								SetTitle(d.innerTitle).
								SetCreatorID(creator.ID).
								SetInviteCode(d.innerInvite).
								Save(innerCtx); err != nil {
								return err
							}
							return d.expectedErr
						})
						if !errors.Is(innerErr, d.expectedErr) {
							return fmt.Errorf("expected inner error %v, got %v", d.expectedErr, innerErr)
						}
						return nil
					})
				}}
			},
			want: func(d *caseData) want {
				return want{
					assert: func(t *testing.T, ctx context.Context, conn *Connection, data *caseData) {
						client := conn.ReadOnlyDB(ctx)
						exists, err := client.User.Query().Where(entuser.UID(data.outerUID)).Exist(ctx)
						if err != nil {
							t.Fatalf("failed to check user existence: %v", err)
						}
						if !exists {
							t.Fatalf("expected outer transaction to commit user")
						}

						pageExists, err := client.Page.Query().Where(entpage.TitleEQ(data.innerTitle)).Exist(ctx)
						if err != nil {
							t.Fatalf("failed to check page existence: %v", err)
						}
						if pageExists {
							t.Fatalf("expected inner page to be rolled back")
						}
					},
				}
			},
		},
	}

	ctx := context.Background()

	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			conn := SetupTestDBConnection(t)
			data := &caseData{}
			if tc.prepare != nil {
				tc.prepare(data)
			}

			args := tc.args(data)
			want := tc.want(data)

			err := args.run(ctx, conn)
			testutil.EqualErr(t, want.err, err)
			if want.assert != nil {
				want.assert(t, ctx, conn, data)
			}
		})
	}
}

type caseData struct {
	uid         string
	title       string
	invite      string
	outerUID    string
	innerTitle  string
	innerInvite string
	expectedErr error
}

func randomInviteCode() string {
	cleaned := strings.ToUpper(strings.ReplaceAll(uuid.NewString(), "-", ""))
	return cleaned[:8]
}
