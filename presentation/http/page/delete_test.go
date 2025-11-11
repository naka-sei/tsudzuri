package page

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/mock/gomock"

	duser "github.com/naka-sei/tsudzuri/domain/user"
	ctxuser "github.com/naka-sei/tsudzuri/pkg/ctx/user"
	"github.com/naka-sei/tsudzuri/pkg/testutil"

	mockdelete "github.com/naka-sei/tsudzuri/usecase/page/mock/mock_delete"
)

func TestDeleteService_Delete(t *testing.T) {
	type mocks struct {
		deleteUsecase *mockdelete.MockDeleteUsecase
	}
	type args struct {
		ctx context.Context
		req DeleteRequest
	}
	type want struct {
		err error
	}

	user := duser.ReconstructUser("user-id-1", "uid-1", "anonymous", nil)

	tests := []struct {
		name  string
		setup func(m *mocks)
		args  args
		want  want
	}{
		{
			name: "success",
			setup: func(m *mocks) {
				m.deleteUsecase.EXPECT().Delete(gomock.Any(), "page-1").Return(nil)
			},
			args: args{
				ctx: ctxuser.WithUser(context.Background(), user),
				req: DeleteRequest{PageID: "page-1"},
			},
			want: want{
				err: nil,
			},
		},
		{
			name: "usecase_error",
			setup: func(m *mocks) {
				m.deleteUsecase.EXPECT().Delete(gomock.Any(), "page-1").Return(errors.New("delete error"))
			},
			args: args{
				ctx: ctxuser.WithUser(context.Background(), user),
				req: DeleteRequest{PageID: "page-1"},
			},
			want: want{
				err: errors.New("delete error"),
			},
		},
		{
			name: "user_not_found",
			setup: func(m *mocks) {
				m.deleteUsecase.EXPECT().Delete(gomock.Any(), "page-1").Return(duser.ErrUserNotFound)
			},
			args: args{
				ctx: ctxuser.WithUser(context.Background(), user),
				req: DeleteRequest{PageID: "page-1"},
			},
			want: want{
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
				deleteUsecase: mockdelete.NewMockDeleteUsecase(ctrl),
			}
			if tt.setup != nil {
				tt.setup(m)
			}
			s := NewDeleteService(m.deleteUsecase)
			err := s.Delete(tt.args.ctx, tt.args.req)
			testutil.EqualErr(t, tt.want.err, err)
		})
	}
}
