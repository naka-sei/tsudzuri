package user

import (
	"context"
	"errors"
	"testing"

	cmp "github.com/google/go-cmp/cmp"
	duser "github.com/naka-sei/tsudzuri/domain/user"
	mockuser "github.com/naka-sei/tsudzuri/domain/user/mock/mock_user"
	"github.com/naka-sei/tsudzuri/pkg/testutil"
	mocktransaction "github.com/naka-sei/tsudzuri/usecase/service/mock/mock_transaction"
	"go.uber.org/mock/gomock"
)

func TestCreateUsecase_Create(t *testing.T) {
	type fields struct {
		userRepo   *mockuser.MockUserRepository
		txnService *mocktransaction.MockTransactionService
	}
	type args struct {
		ctx context.Context
		uid string
	}
	type want struct {
		user *duser.User
		err  error
	}

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
				user := duser.NewUser("uid-1")
				f.userRepo.EXPECT().Save(gomock.Any(), user).Return(nil)
			},
			args: args{
				ctx: context.Background(),
				uid: "uid-1",
			},
			want: want{
				user: func() *duser.User {
					u := duser.NewUser("uid-1")
					return u
				}(),
				err: nil,
			},
		},
		{
			name: "user_save_error",
			setup: func(f *fields) {
				f.txnService.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					},
				)
				user := duser.NewUser("fail-uid")
				f.userRepo.EXPECT().Save(gomock.Any(), user).Return(errors.New("save error"))
			},
			args: args{
				ctx: context.Background(),
				uid: "fail-uid",
			},
			want: want{
				user: nil,
				err:  errors.New("save error"),
			},
		},
		{
			name: "transaction_error",
			setup: func(f *fields) {
				f.txnService.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Return(errors.New("txn error"))
			},
			args: args{
				ctx: context.Background(),
				uid: "txn-uid",
			},
			want: want{
				user: nil,
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
				userRepo:   mockuser.NewMockUserRepository(ctrl),
				txnService: mocktransaction.NewMockTransactionService(ctrl),
			}
			if tt.setup != nil {
				tt.setup(f)
			}
			u := NewCreateUsecase(f.userRepo, f.txnService)
			got, err := u.Create(tt.args.ctx, tt.args.uid)
			if diff := cmp.Diff(tt.want.user, got, cmp.AllowUnexported(duser.User{})); diff != "" {
				t.Errorf("user mismatch (-want +got):\n%s", diff)
			}
			testutil.EqualErr(t, tt.want.err, err)
		})
	}
}
