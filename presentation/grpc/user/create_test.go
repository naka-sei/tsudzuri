package user

import (
	"context"
	"errors"
	"testing"
	"time"

	cmp "github.com/google/go-cmp/cmp"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/testing/protocmp"

	tsudzuriv1 "github.com/naka-sei/tsudzuri/api/tsudzuri/v1"
	duser "github.com/naka-sei/tsudzuri/domain/user"
	"github.com/naka-sei/tsudzuri/pkg/cache"
	"github.com/naka-sei/tsudzuri/pkg/testutil"
	mockcreate "github.com/naka-sei/tsudzuri/usecase/user/mock/mock_create"
)

func TestCreateService_Create(t *testing.T) {
	type args struct {
		ctx context.Context
		req *tsudzuriv1.CreateUserRequest
	}
	type want struct {
		res *tsudzuriv1.User
		err error
	}
	type verify func(t *testing.T, c cache.Cache[*duser.User])

	tests := []struct {
		name  string
		setup func(m *mockcreate.MockCreateUsecase)
		args  args
		want  want
		do    verify
	}{
		{
			name: "success",
			setup: func(m *mockcreate.MockCreateUsecase) {
				u := duser.ReconstructUser("id-1", "uid-1", "anonymous", nil)
				m.EXPECT().Create(gomock.Any(), "uid-1").Return(u, nil)
			},
			args: args{
				ctx: context.Background(),
				req: &tsudzuriv1.CreateUserRequest{Uid: "uid-1"},
			},
			want: want{
				res: &tsudzuriv1.User{Id: "id-1", Uid: "uid-1", Provider: "anonymous"},
				err: nil,
			},
			do: func(t *testing.T, c cache.Cache[*duser.User]) {
				t.Helper()
				cached, ok := c.Get(context.Background(), "uid-1")
				if !ok {
					t.Fatalf("expected user cached")
				}
				if diff := cmp.Diff("uid-1", cached.UID()); diff != "" {
					t.Fatalf("cached uid mismatch (-want +got):\n%s", diff)
				}
			},
		},
		{
			name: "usecase_error",
			setup: func(m *mockcreate.MockCreateUsecase) {
				m.EXPECT().Create(gomock.Any(), "uid-err").Return(nil, errors.New("usecase error"))
			},
			args: args{
				ctx: context.Background(),
				req: &tsudzuriv1.CreateUserRequest{Uid: "uid-err"},
			},
			want: want{
				res: nil,
				err: errors.New("usecase error"),
			},
		},
		{
			name: "usecase_returns_nil_user",
			setup: func(m *mockcreate.MockCreateUsecase) {
				m.EXPECT().Create(gomock.Any(), "uid-nil").Return(nil, nil)
			},
			args: args{
				ctx: context.Background(),
				req: &tsudzuriv1.CreateUserRequest{Uid: "uid-nil"},
			},
			want: want{
				res: nil,
				err: nil,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			usecase := mockcreate.NewMockCreateUsecase(ctrl)
			if tt.setup != nil {
				tt.setup(usecase)
			}

			svc := NewCreateService(usecase)
			userCache := cache.NewMemoryCache[*duser.User](time.Minute)
			svc.SetCache(userCache)

			got, err := svc.Create(tt.args.ctx, tt.args.req)
			if diff := cmp.Diff(tt.want.res, got, protocmp.Transform()); diff != "" {
				t.Fatalf("response mismatch (-want +got):\n%s", diff)
			}
			testutil.EqualErr(t, tt.want.err, err)
			if tt.do != nil {
				tt.do(t, userCache)
			}
		})
	}
}
