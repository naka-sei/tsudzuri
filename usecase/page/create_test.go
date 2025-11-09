package page

import (
	"context"
	"errors"
	"testing"

	cmp "github.com/google/go-cmp/cmp"
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
				page, _ := dpage.NewPage("test-title", user)
				f.pageRepo.EXPECT().Save(gomock.Any(), page).Return(nil)
			},
			args: args{
				ctx:   ctxuser.WithUser(context.Background(), user),
				title: "test-title",
			},
			want: want{
				page: func() *dpage.Page {
					p, _ := dpage.NewPage("test-title", user)
					return p
				}(),
				err: nil,
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
				page, _ := dpage.NewPage("fail-title", user)
				f.pageRepo.EXPECT().Save(gomock.Any(), page).Return(errors.New("save error"))
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
			if diff := cmp.Diff(tt.want.page, got, cmp.AllowUnexported(dpage.Page{}, duser.User{})); diff != "" {
				t.Errorf("page mismatch (-want +got):\n%s", diff)
			}
			testutil.EqualErr(t, tt.want.err, err)
		})
	}
}
