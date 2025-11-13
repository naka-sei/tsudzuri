package page

import (
	"context"
	"errors"
	"testing"

	cmp "github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"go.uber.org/mock/gomock"

	dpage "github.com/naka-sei/tsudzuri/domain/page"
	mockpage "github.com/naka-sei/tsudzuri/domain/page/mock/mock_page"
	duser "github.com/naka-sei/tsudzuri/domain/user"
	ctxuser "github.com/naka-sei/tsudzuri/pkg/ctx/user"
	"github.com/naka-sei/tsudzuri/pkg/testutil"
	mocktransaction "github.com/naka-sei/tsudzuri/usecase/service/mock/mock_transaction"
)

func TestCreateUsecase_Create(t *testing.T) {
	type fields struct {
		pageRepo   *mockpage.MockPageRepository
		txnService *mocktransaction.MockTransactionService
	}
	type args struct {
		ctx   context.Context
		title string
	}
	type want struct {
		page *dpage.Page
		err  error
	}

	user := duser.NewUser("uid-1")

	tests := []struct {
		name  string
		setup func(f *fields)
		args  args
		want  want
	}{
		{
			name: "success",
			setup: func(f *fields) {
				f.txnService.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					},
				)
				f.pageRepo.EXPECT().Save(gomock.Any(), gomock.AssignableToTypeOf(&dpage.Page{})).DoAndReturn(
					func(ctx context.Context, pg *dpage.Page) (*dpage.Page, error) {
						if pg.Title() != "test-title" {
							t.Fatalf("unexpected title: %s", pg.Title())
						}
						return pg, nil
					},
				)
			},
			args: args{
				ctx:   ctxuser.WithUser(context.Background(), user),
				title: "test-title",
			},
			want: want{
				page: dpage.ReconstructPage("", "test-title", *user, "", dpage.Links{}, nil),
				err:  nil,
			},
		},
		{
			name: "page_save_error",
			setup: func(f *fields) {
				f.txnService.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					},
				)
				f.pageRepo.EXPECT().Save(gomock.Any(), gomock.AssignableToTypeOf(&dpage.Page{})).Return(nil, errors.New("save error"))
			},
			args: args{
				ctx:   ctxuser.WithUser(context.Background(), user),
				title: "fail-title",
			},
			want: want{
				page: nil,
				err:  errors.New("save error"),
			},
		},
		{
			name: "transaction_error",
			setup: func(f *fields) {
				f.txnService.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Return(errors.New("txn error"))
			},
			args: args{
				ctx:   ctxuser.WithUser(context.Background(), user),
				title: "txn-title",
			},
			want: want{
				page: nil,
				err:  errors.New("txn error"),
			},
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f := &fields{
				pageRepo:   mockpage.NewMockPageRepository(ctrl),
				txnService: mocktransaction.NewMockTransactionService(ctrl),
			}
			if tt.setup != nil {
				tt.setup(f)
			}
			u := NewCreateUsecase(f.pageRepo, f.txnService)
			got, err := u.Create(tt.args.ctx, tt.args.title)
			if tt.want.page != nil {
				if got == nil {
					t.Fatalf("expected page, got nil")
				}
				if diff := cmp.Diff(tt.want.page, got, cmp.AllowUnexported(dpage.Page{}, duser.User{}), cmpopts.IgnoreFields(dpage.Page{}, "inviteCode")); diff != "" {
					t.Errorf("page mismatch (-want +got):\n%s", diff)
				}
			} else if got != nil {
				t.Errorf("expected nil page, got %+v", got)
			}
			testutil.EqualErr(t, tt.want.err, err)
		})
	}
}
