package user

import (
	"context"
	"errors"
	"testing"

	cmp "github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/emptypb"

	tsudzuriv1 "github.com/naka-sei/tsudzuri/api/tsudzuri/v1"
	duser "github.com/naka-sei/tsudzuri/domain/user"
	ctxuser "github.com/naka-sei/tsudzuri/pkg/ctx/user"
	"github.com/naka-sei/tsudzuri/pkg/testutil"
	mockget "github.com/naka-sei/tsudzuri/usecase/user/mock/mock_get"
	"go.uber.org/mock/gomock"
)

func TestGetService_Get(t *testing.T) {
	type fields struct {
		usecase *mockget.MockGetUsecase
	}
	type args struct {
		ctx context.Context
		req *emptypb.Empty
	}
	type want struct {
		res *tsudzuriv1.User
		err error
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
				f.usecase.EXPECT().Get(gomock.Any()).Return(duser.ReconstructUser("id-1", "uid-1", string(duser.ProviderGoogle), nil, duser.WithJoinedPageIDs([]string{"page-1", "page-2"})), nil)
			},
			args: args{
				ctx: ctxuser.WithUser(context.Background(), duser.ReconstructUser("id-ctx", "uid-1", string(duser.ProviderGoogle), nil)),
				req: &emptypb.Empty{},
			},
			want: want{
				res: &tsudzuriv1.User{
					Id:            "id-1",
					Uid:           "uid-1",
					Provider:      string(duser.ProviderGoogle),
					JoinedPageIds: []string{"page-1", "page-2"},
				},
			},
		},
		{
			name: "usecase_error",
			setup: func(f *fields) {
				f.usecase.EXPECT().Get(gomock.Any()).Return(nil, errors.New("usecase error"))
			},
			args: args{
				ctx: ctxuser.WithUser(context.Background(), duser.ReconstructUser("id-ctx", "uid-err", string(duser.ProviderGoogle), nil)),
				req: &emptypb.Empty{},
			},
			want: want{
				err: errors.New("usecase error"),
			},
		},
		{
			name:  "no_user_in_context",
			setup: func(f *fields) {},
			args: args{
				ctx: context.Background(),
				req: &emptypb.Empty{},
			},
			want: want{
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

			f := &fields{usecase: mockget.NewMockGetUsecase(ctrl)}
			if tt.setup != nil {
				tt.setup(f)
			}

			svc := NewGetService(f.usecase)
			got, err := svc.Get(tt.args.ctx, tt.args.req)

			if diff := cmp.Diff(tt.want.res, got, protocmp.Transform()); diff != "" {
				t.Fatalf("response mismatch (-want +got):\n%s", diff)
			}
			testutil.EqualErr(t, tt.want.err, err)
		})
	}
}
