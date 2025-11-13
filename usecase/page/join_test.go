package page

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/mock/gomock"

	dpage "github.com/naka-sei/tsudzuri/domain/page"
	mockpage "github.com/naka-sei/tsudzuri/domain/page/mock/mock_page"
	duser "github.com/naka-sei/tsudzuri/domain/user"
	ctxuser "github.com/naka-sei/tsudzuri/pkg/ctx/user"
	"github.com/naka-sei/tsudzuri/pkg/testutil"
	mocktransaction "github.com/naka-sei/tsudzuri/usecase/service/mock/mock_transaction"
)

func TestJoinUsecase_Join(t *testing.T) {
	type fields struct {
		pageRepo   *mockpage.MockPageRepository
		txnService *mocktransaction.MockTransactionService
	}
	type args struct {
		ctx        context.Context
		pageID     string
		inviteCode string
	}
	type want struct {
		err error
	}

	newCtxWithUser := func(id string) (context.Context, *duser.User) {
		user := duser.ReconstructUser(id, "uid-"+id, "anonymous", nil)
		return ctxuser.WithUser(context.Background(), user), user
	}

	tests := []struct {
		name  string
		setup func(t *testing.T, f *fields, tt *args)
		args  args
		want  want
	}{
		{
			name: "success",
			args: func() args {
				ctx, _ := newCtxWithUser("user-id")
				return args{
					ctx:        ctx,
					pageID:     "page-id",
					inviteCode: "INVITE01",
				}
			}(),
			setup: func(t *testing.T, f *fields, tt *args) {
				creator := duser.ReconstructUser("creator-id", "uid-creator", "anonymous", nil)
				page := dpage.ReconstructPage(tt.pageID, "Title", *creator, tt.inviteCode, dpage.Links{}, duser.Users{})

				f.pageRepo.EXPECT().Get(gomock.Any(), tt.pageID).Return(page, nil)
				f.txnService.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					},
				)
				f.pageRepo.EXPECT().Save(gomock.Any(), page).DoAndReturn(
					func(ctx context.Context, pg *dpage.Page) (*dpage.Page, error) {
						if len(pg.InvitedUsers()) != 1 {
							t.Fatalf("expected invited users to have 1 entry, got %d", len(pg.InvitedUsers()))
						}
						if pg.InvitedUsers()[0].ID() != "user-id" {
							t.Fatalf("unexpected user id: %s", pg.InvitedUsers()[0].ID())
						}
						return pg, nil
					},
				)
			},
			want: want{err: nil},
		},
		{
			name: "user_not_found_in_context",
			args: args{
				ctx:        context.Background(),
				pageID:     "page-id",
				inviteCode: "INVITE01",
			},
			setup: func(t *testing.T, f *fields, tt *args) {},
			want:  want{err: duser.ErrUserNotFound},
		},
		{
			name: "page_not_found",
			args: func() args {
				ctx, _ := newCtxWithUser("user-id")
				return args{ctx: ctx, pageID: "missing", inviteCode: "INVITE01"}
			}(),
			setup: func(t *testing.T, f *fields, tt *args) {
				f.pageRepo.EXPECT().Get(gomock.Any(), tt.pageID).Return(nil, nil)
			},
			want: want{err: ErrPageNotFound},
		},
		{
			name: "get_error",
			args: func() args {
				ctx, _ := newCtxWithUser("user-id")
				return args{ctx: ctx, pageID: "page-id", inviteCode: "INVITE01"}
			}(),
			setup: func(t *testing.T, f *fields, tt *args) {
				f.pageRepo.EXPECT().Get(gomock.Any(), tt.pageID).Return(nil, errors.New("get error"))
			},
			want: want{err: errors.New("get error")},
		},
		{
			name: "invalid_invite_code",
			args: func() args {
				ctx, _ := newCtxWithUser("user-id")
				return args{ctx: ctx, pageID: "page-id", inviteCode: "WRONG"}
			}(),
			setup: func(t *testing.T, f *fields, tt *args) {
				creator := duser.ReconstructUser("creator-id", "uid-creator", "anonymous", nil)
				page := dpage.ReconstructPage(tt.pageID, "Title", *creator, "INVITE01", dpage.Links{}, duser.Users{})

				f.pageRepo.EXPECT().Get(gomock.Any(), tt.pageID).Return(page, nil)
				f.txnService.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					},
				)
			},
			want: want{err: dpage.ErrInvalidInviteCode},
		},
		{
			name: "save_error",
			args: func() args {
				ctx, _ := newCtxWithUser("user-id")
				return args{ctx: ctx, pageID: "page-id", inviteCode: "INVITE01"}
			}(),
			setup: func(t *testing.T, f *fields, tt *args) {
				creator := duser.ReconstructUser("creator-id", "uid-creator", "anonymous", nil)
				page := dpage.ReconstructPage(tt.pageID, "Title", *creator, tt.inviteCode, dpage.Links{}, duser.Users{})

				f.pageRepo.EXPECT().Get(gomock.Any(), tt.pageID).Return(page, nil)
				f.txnService.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					},
				)
				f.pageRepo.EXPECT().Save(gomock.Any(), page).Return(nil, errors.New("save error"))
			},
			want: want{err: errors.New("save error")},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			f := &fields{
				pageRepo:   mockpage.NewMockPageRepository(ctrl),
				txnService: mocktransaction.NewMockTransactionService(ctrl),
			}
			if tt.setup != nil {
				tt.setup(t, f, &tt.args)
			}

			u := NewJoinUsecase(f.pageRepo, f.txnService)
			err := u.Join(tt.args.ctx, tt.args.pageID, tt.args.inviteCode)
			testutil.EqualErr(t, tt.want.err, err)
		})
	}
}
