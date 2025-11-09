package page

import (
	"context"
	"errors"
	"testing"

	cmp "github.com/google/go-cmp/cmp"
	"go.uber.org/mock/gomock"

	dpage "github.com/naka-sei/tsudzuri/domain/page"
	duser "github.com/naka-sei/tsudzuri/domain/user"
	ctxuser "github.com/naka-sei/tsudzuri/pkg/ctx/user"
	"github.com/naka-sei/tsudzuri/pkg/testutil"
	"github.com/naka-sei/tsudzuri/presentation/http/response"
	mockcreate "github.com/naka-sei/tsudzuri/usecase/page/mock/mock_create"
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
		res response.EmptyResponse
		err error
	}

	user := duser.NewUser("uid-1")

	tests := []struct {
		name  string
		setup func(m *mocks)
		args  args
		want  want
	}{
		{
			name: "success",
			setup: func(m *mocks) {
				page, _ := dpage.NewPage("test-title", user)
				m.createUsecase.EXPECT().Create(gomock.Any(), "test-title").Return(page, nil)
			},
			args: args{
				ctx: ctxuser.WithUser(context.Background(), user),
				req: CreateRequest{Title: "test-title"},
			},
			want: want{
				res: response.EmptyResponse{},
				err: nil,
			},
		},
		{
			name: "usecase_error",
			setup: func(m *mocks) {
				m.createUsecase.EXPECT().Create(gomock.Any(), "fail-title").Return(nil, errors.New("usecase error"))
			},
			args: args{
				ctx: ctxuser.WithUser(context.Background(), user),
				req: CreateRequest{Title: "fail-title"},
			},
			want: want{
				res: response.EmptyResponse{},
				err: errors.New("usecase error"),
			},
		},
		{
			name: "user_not_found",
			setup: func(m *mocks) {
				// No expectation since user check happens before usecase call
			},
			args: args{
				ctx: context.Background(),
				req: CreateRequest{Title: "test-title"},
			},
			want: want{
				res: response.EmptyResponse{},
				err: duser.ErrUserNotFound,
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
			got, err := s.Create(tt.args.ctx, tt.args.req)
			if diff := cmp.Diff(tt.want.res, got); diff != "" {
				t.Errorf("response mismatch (-want +got):\n%s", diff)
			}
			testutil.EqualErr(t, tt.want.err, err)
		})
	}
}
