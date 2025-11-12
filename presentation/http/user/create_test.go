package user

import (
	"context"
	"errors"
	"testing"
	"time"

	cmp "github.com/google/go-cmp/cmp"
	"go.uber.org/mock/gomock"

	duser "github.com/naka-sei/tsudzuri/domain/user"
	"github.com/naka-sei/tsudzuri/pkg/cache"
	"github.com/naka-sei/tsudzuri/pkg/testutil"
	mockcreate "github.com/naka-sei/tsudzuri/usecase/user/mock/mock_create"
)

func TestCreateService_Create(t *testing.T) {
	type mocks struct {
		createUsecase *mockcreate.MockCreateUsecase
	}
	type args struct {
		ctx context.Context
		req CreateRequest
	}
	type want struct {
		res *UserResponse
		err error
	}
	type verify func(t *testing.T, c cache.Cache[*duser.User])

	tests := []struct {
		name  string
		setup func(m *mocks)
		args  args
		want  want
		do    verify
	}{
		{
			name: "success",
			setup: func(m *mocks) {
				user := duser.ReconstructUser("id-1", "test-uid", "anonymous", nil)
				m.createUsecase.EXPECT().Create(gomock.Any(), "test-uid").Return(user, nil)
			},
			args: args{
				ctx: context.Background(),
				req: CreateRequest{UID: "test-uid"},
			},
			want: want{
				res: &UserResponse{
					ID:       "id-1",
					UID:      "test-uid",
					Provider: "anonymous",
					Email:    nil,
				},
				err: nil,
			},
			do: func(t *testing.T, c cache.Cache[*duser.User]) {
				t.Helper()
				cached, ok := c.Get(context.Background(), "test-uid")
				if !ok {
					t.Fatalf("expected user cached")
				}
				if diff := cmp.Diff("test-uid", cached.UID()); diff != "" {
					t.Fatalf("cached uid mismatch (-want +got):\n%s", diff)
				}
			},
		},
		{
			name: "usecase_error",
			setup: func(m *mocks) {
				m.createUsecase.EXPECT().Create(gomock.Any(), "fail-uid").Return(nil, errors.New("usecase error"))
			},
			args: args{
				ctx: context.Background(),
				req: CreateRequest{UID: "fail-uid"},
			},
			want: want{
				res: nil,
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
				createUsecase: mockcreate.NewMockCreateUsecase(ctrl),
			}
			if tt.setup != nil {
				tt.setup(m)
			}
			s := NewCreateService(m.createUsecase)
			userCache := cache.NewMemoryCache[*duser.User](time.Minute)
			s.SetCache(userCache)
			got, err := s.Create(tt.args.ctx, tt.args.req)
			if diff := cmp.Diff(tt.want.res, got); diff != "" {
				t.Errorf("response mismatch (-want +got):\n%s", diff)
			}
			testutil.EqualErr(t, tt.want.err, err)
			if tt.do != nil {
				tt.do(t, userCache)
			}
		})
	}
}
