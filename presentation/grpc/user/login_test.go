package user

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	tsudzuriv1 "github.com/naka-sei/tsudzuri/api/tsudzuri/v1"
	duser "github.com/naka-sei/tsudzuri/domain/user"
	ctxuser "github.com/naka-sei/tsudzuri/pkg/ctx/user"
	"github.com/naka-sei/tsudzuri/pkg/testutil"
	mocklogin "github.com/naka-sei/tsudzuri/usecase/user/mock/mock_login"
)

func TestLoginService_Login(t *testing.T) {
	type args struct {
		ctx context.Context
		req *tsudzuriv1.LoginRequest
	}
	type want struct {
		res *emptypb.Empty
		err error
	}

	user := duser.ReconstructUser("id-1", "uid-1", "anonymous", nil)

	tests := []struct {
		name  string
		setup func(t *testing.T, m *mocklogin.MockLoginUsecase)
		args  args
		want  want
	}{
		{
			name: "success",
			setup: func(t *testing.T, m *mocklogin.MockLoginUsecase) {
				m.EXPECT().Login(gomock.Any(), "google", gomock.Any()).DoAndReturn(
					func(_ context.Context, provider string, email *string) error {
						t.Helper()
						if email == nil || *email != "user@example.com" {
							t.Fatalf("unexpected email pointer: %v", email)
						}
						return nil
					},
				)
			},
			args: args{
				ctx: ctxuser.WithUser(context.Background(), user),
				req: &tsudzuriv1.LoginRequest{
					Provider: "google",
					Email:    wrapperspb.String("user@example.com"),
				},
			},
			want: want{
				res: &emptypb.Empty{},
				err: nil,
			},
		},
		{
			name: "usecase_error",
			setup: func(t *testing.T, m *mocklogin.MockLoginUsecase) {
				m.EXPECT().Login(gomock.Any(), "google", gomock.Any()).DoAndReturn(
					func(_ context.Context, provider string, email *string) error {
						t.Helper()
						if email == nil || *email != "fail@example.com" {
							t.Fatalf("unexpected email pointer: %v", email)
						}
						return errors.New("login error")
					},
				)
			},
			args: args{
				ctx: ctxuser.WithUser(context.Background(), user),
				req: &tsudzuriv1.LoginRequest{
					Provider: "google",
					Email:    wrapperspb.String("fail@example.com"),
				},
			},
			want: want{
				res: nil,
				err: errors.New("login error"),
			},
		},
		{
			name:  "user_not_found",
			setup: nil,
			args: args{
				ctx: context.Background(),
				req: &tsudzuriv1.LoginRequest{Provider: "google"},
			},
			want: want{
				res: nil,
				err: duser.ErrUserNotFound,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			usecase := mocklogin.NewMockLoginUsecase(ctrl)
			if tt.setup != nil {
				tt.setup(t, usecase)
			}

			svc := NewLoginService(usecase)
			got, err := svc.Login(tt.args.ctx, tt.args.req)
			if err == nil && tt.want.err == nil && got == nil {
				t.Fatalf("expected non-nil response when no error")
			}
			testutil.EqualErr(t, tt.want.err, err)
		})
	}
}
