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
	dpage "github.com/naka-sei/tsudzuri/domain/page"
	duser "github.com/naka-sei/tsudzuri/domain/user"
	ctxuser "github.com/naka-sei/tsudzuri/pkg/ctx/user"
	"github.com/naka-sei/tsudzuri/pkg/testutil"
	mockcreate "github.com/naka-sei/tsudzuri/usecase/page/mock/mock_create"
)

func TestCreateService_Create(t *testing.T) {
	type args struct {
		ctx context.Context
		req *tsudzuriv1.CreatePageRequest
	}
	type want struct {
		res *emptypb.Empty
		err error
	}

	user := duser.ReconstructUser("user-id-1", "uid-1", "anonymous", nil)

	tests := []struct {
		name  string
		setup func(m *mockcreate.MockCreateUsecase)
		args  args
		want  want
	}{
		{
			name: "success",
			setup: func(m *mockcreate.MockCreateUsecase) {
				page := dpage.ReconstructPage("page-id", "test-title", *user, "invite", nil, nil)
				m.EXPECT().Create(gomock.Any(), "test-title").Return(page, nil)
			},
			args: args{
				ctx: ctxuser.WithUser(context.Background(), user),
				req: &tsudzuriv1.CreatePageRequest{Title: "test-title"},
			},
			want: want{
				res: &emptypb.Empty{},
				err: nil,
			},
		},
		{
			name: "usecase_error",
			setup: func(m *mockcreate.MockCreateUsecase) {
				m.EXPECT().Create(gomock.Any(), "fail-title").Return(nil, errors.New("usecase error"))
			},
			args: args{
				ctx: ctxuser.WithUser(context.Background(), user),
				req: &tsudzuriv1.CreatePageRequest{Title: "fail-title"},
			},
			want: want{
				res: nil,
				err: errors.New("usecase error"),
			},
		},
		{
			name:  "user_not_found",
			setup: nil,
			args: args{
				ctx: context.Background(),
				req: &tsudzuriv1.CreatePageRequest{Title: "test-title"},
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

			usecase := mockcreate.NewMockCreateUsecase(ctrl)
			if tt.setup != nil {
				tt.setup(usecase)
			}

			svc := NewCreateService(usecase)
			got, err := svc.Create(tt.args.ctx, tt.args.req)
			if diff := cmp.Diff(tt.want.res, got, protocmp.Transform()); diff != "" {
				t.Fatalf("response mismatch (-want +got):\n%s", diff)
			}
			testutil.EqualErr(t, tt.want.err, err)
		})
	}
}
