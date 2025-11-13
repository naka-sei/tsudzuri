package page

import (
	"context"
	"errors"
	"testing"

	cmp "github.com/google/go-cmp/cmp"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/emptypb"

	tsudzuriv1 "github.com/naka-sei/tsudzuri/api/tsudzuri/v1"
	duser "github.com/naka-sei/tsudzuri/domain/user"
	ctxuser "github.com/naka-sei/tsudzuri/pkg/ctx/user"
	"github.com/naka-sei/tsudzuri/pkg/testutil"
	mockdelete "github.com/naka-sei/tsudzuri/usecase/page/mock/mock_delete"
)

func TestDeleteService_Delete(t *testing.T) {
	type args struct {
		ctx context.Context
		req *tsudzuriv1.DeletePageRequest
	}
	type want struct {
		res *emptypb.Empty
		err error
	}

	user := duser.ReconstructUser("user-id", "uid-1", "anonymous", nil)

	tests := []struct {
		name  string
		setup func(m *mockdelete.MockDeleteUsecase)
		args  args
		want  want
	}{
		{
			name: "success",
			setup: func(m *mockdelete.MockDeleteUsecase) {
				m.EXPECT().Delete(gomock.Any(), "page-1").Return(nil)
			},
			args: args{
				ctx: ctxuser.WithUser(context.Background(), user),
				req: &tsudzuriv1.DeletePageRequest{PageId: "page-1"},
			},
			want: want{
				res: &emptypb.Empty{},
				err: nil,
			},
		},
		{
			name: "usecase_error",
			setup: func(m *mockdelete.MockDeleteUsecase) {
				m.EXPECT().Delete(gomock.Any(), "page-1").Return(errors.New("delete error"))
			},
			args: args{
				ctx: ctxuser.WithUser(context.Background(), user),
				req: &tsudzuriv1.DeletePageRequest{PageId: "page-1"},
			},
			want: want{
				res: nil,
				err: errors.New("delete error"),
			},
		},
		{
			name:  "user_not_found",
			setup: nil,
			args: args{
				ctx: context.Background(),
				req: &tsudzuriv1.DeletePageRequest{PageId: "page-1"},
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

			usecase := mockdelete.NewMockDeleteUsecase(ctrl)
			if tt.setup != nil {
				tt.setup(usecase)
			}

			svc := NewDeleteService(usecase)
			got, err := svc.Delete(tt.args.ctx, tt.args.req)
			if diff := cmp.Diff(tt.want.res, got, protocmp.Transform()); diff != "" {
				t.Fatalf("response mismatch (-want +got):\n%s", diff)
			}
			testutil.EqualErr(t, tt.want.err, err)
		})
	}
}
