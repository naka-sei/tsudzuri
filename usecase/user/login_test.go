package user

import (
	"context"
	"errors"
	"testing"

	cmp "github.com/google/go-cmp/cmp"
	duser "github.com/naka-sei/tsudzuri/domain/user"
	mockuser "github.com/naka-sei/tsudzuri/domain/user/mock/mock_user"
	ctxuser "github.com/naka-sei/tsudzuri/pkg/ctx/user"
	"github.com/naka-sei/tsudzuri/pkg/ptr"
	"github.com/naka-sei/tsudzuri/pkg/testutil"
	mocktransaction "github.com/naka-sei/tsudzuri/usecase/service/mock/mock_transaction"
	"go.uber.org/mock/gomock"
)

func TestLoginUsecase_Login(t *testing.T) {
	type fields struct {
		userRepo   *mockuser.MockUserRepository
		txnService *mocktransaction.MockTransactionService
	}
	type args struct {
		ctx      context.Context
		provider string
		email    *string
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
				f.userRepo.EXPECT().Save(gomock.Any(), gomock.AssignableToTypeOf(&duser.User{})).Return(nil, nil)
			},
			args: args{
				ctx:      ctxuser.WithUser(context.Background(), duser.NewUser("uid-1")),
				provider: "google",
				email:    ptr.Ptr("u@example.com"),
			},
			want: want{
				user: func() *duser.User {
					return duser.ReconstructUser("", "uid-1", "google", ptr.Ptr("u@example.com"))
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
				f.userRepo.EXPECT().Save(gomock.Any(), gomock.AssignableToTypeOf(&duser.User{})).Return(nil, errors.New("save error"))
			},
			args: args{
				ctx:      ctxuser.WithUser(context.Background(), duser.NewUser("fail-uid")),
				provider: "google",
				email:    ptr.Ptr("u@example.com"),
			},
			want: want{
				user: func() *duser.User {
					return duser.ReconstructUser("", "fail-uid", "google", ptr.Ptr("u@example.com"))
				}(),
				err: errors.New("save error"),
			},
		},
		{
			name: "transaction_error",
			setup: func(f *fields) {
				f.txnService.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Return(errors.New("txn error"))
			},
			args: args{
				ctx:      ctxuser.WithUser(context.Background(), duser.NewUser("txn-uid")),
				provider: "google",
				email:    ptr.Ptr("u@example.com"),
			},
			want: want{
				user: func() *duser.User {
					return duser.ReconstructUser("", "txn-uid", "google", ptr.Ptr("u@example.com"))
				}(),
				err: errors.New("txn error"),
			},
		},
		{
			name:  "no_user_in_context",
			setup: func(f *fields) {},
			args: args{
				ctx:      context.Background(),
				provider: "google",
				email:    ptr.Ptr("u@example.com"),
			},
			want: want{
				user: nil,
				err:  duser.ErrUserNotFound,
			},
		},
		{
			name:  "no_email",
			setup: func(f *fields) {},
			args: args{
				ctx:      ctxuser.WithUser(context.Background(), duser.NewUser("uid-1")),
				provider: "google",
				email:    nil,
			},
			want: want{
				user: nil,
				err:  duser.ErrNoSpecifiedEmail,
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

			u := NewLoginUsecase(f.userRepo, f.txnService)
			err := u.Login(tt.args.ctx, tt.args.provider, tt.args.email)

			if gotUser, ok := ctxuser.UserFromContext(tt.args.ctx); ok && tt.want.user != nil {
				if diff := cmp.Diff(tt.want.user, gotUser, cmp.AllowUnexported(duser.User{})); diff != "" {
					t.Errorf("user mismatch (-want +got):\n%s", diff)
				}
			}

			testutil.EqualErr(t, tt.want.err, err)
		})
	}
}
