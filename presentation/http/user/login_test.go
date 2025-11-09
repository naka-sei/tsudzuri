package user

import (
	"context"
	"errors"
	"testing"

	cmp "github.com/google/go-cmp/cmp"
	"go.uber.org/mock/gomock"

	duser "github.com/naka-sei/tsudzuri/domain/user"
	ctxuser "github.com/naka-sei/tsudzuri/pkg/ctx/user"
	"github.com/naka-sei/tsudzuri/pkg/ptr"
	"github.com/naka-sei/tsudzuri/pkg/testutil"
	"github.com/naka-sei/tsudzuri/presentation/http/response"
	mocklogin "github.com/naka-sei/tsudzuri/usecase/user/mock/mock_login"
)

func TestLoginService_Login(t *testing.T) {
	type mocks struct {
		loginUsecase *mocklogin.MockLoginUsecase
	}
	type args struct {
		ctx context.Context
		req LoginRequest
	}
	type want struct {
		res response.EmptyResponse
		err error
	}

	user := duser.ReconstructUser("id-1", "uid-1", "anonymous", nil)

	tests := []struct {
		name  string
		setup func(m *mocks)
		args  args
		want  want
	}{
		{
			name: "success",
			setup: func(m *mocks) {
				m.loginUsecase.EXPECT().Login(gomock.Any(), "google", ptr.Ptr("u@example.com")).Return(nil)
			},
			args: args{
				ctx: ctxuser.WithUser(context.Background(), user),
				req: LoginRequest{Provider: "google", Email: ptr.Ptr("u@example.com")},
			},
			want: want{
				res: response.EmptyResponse{},
				err: nil,
			},
		},
		{
			name: "usecase_error",
			setup: func(m *mocks) {
				m.loginUsecase.EXPECT().Login(gomock.Any(), "google", ptr.Ptr("fail@example.com")).Return(errors.New("usecase error"))
			},
			args: args{
				ctx: ctxuser.WithUser(context.Background(), user),
				req: LoginRequest{Provider: "google", Email: ptr.Ptr("fail@example.com")},
			},
			want: want{
				res: response.EmptyResponse{},
				err: errors.New("usecase error"),
			},
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := &mocks{
				loginUsecase: mocklogin.NewMockLoginUsecase(ctrl),
			}
			if tt.setup != nil {
				tt.setup(m)
			}
			s := NewLoginService(m.loginUsecase)
			got, err := s.Login(tt.args.ctx, tt.args.req)
			if diff := cmp.Diff(tt.want.res, got); diff != "" {
				t.Errorf("response mismatch (-want +got):\n%s", diff)
			}
			testutil.EqualErr(t, tt.want.err, err)
		})
	}
}
