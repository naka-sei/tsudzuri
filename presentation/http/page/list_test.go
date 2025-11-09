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
	mocklist "github.com/naka-sei/tsudzuri/usecase/page/mock/mock_list"
)

func TestListService_List(t *testing.T) {
	type mocks struct {
		listUsecase *mocklist.MockListUsecase
	}
	type args struct {
		ctx context.Context
		req ListRequest
	}
	type want struct {
		res []PageResponse
		err error
	}

	creator := duser.ReconstructUser("user-id-1", "uid-1", "anonymous", nil)
	invitedUser := duser.ReconstructUser("user-id-2", "uid-2", "invited", nil)
	otherUser := duser.ReconstructUser("user-id-3", "uid-3", "other", nil)

	p1 := dpage.ReconstructPage("page-1", "t1", *creator, "invite-1", dpage.Links{}, duser.Users{})
	p2 := dpage.ReconstructPage("page-2", "t2", *creator, "invite-2", dpage.Links{dpage.ReconstructLink("url1", "memo1", 1)}, duser.Users{})
	p3 := dpage.ReconstructPage("page-3", "t3", *otherUser, "invite-3", dpage.Links{}, duser.Users{invitedUser})

	tests := []struct {
		name  string
		setup func(m *mocks)
		args  args
		want  want
	}{
		{
			name: "success_list_by_creator_user",
			setup: func(m *mocks) {
				m.listUsecase.EXPECT().List(gomock.Any()).Return([]*dpage.Page{p1, p2}, nil)
			},
			args: args{
				ctx: ctxuser.WithUser(context.Background(), creator),
				req: ListRequest{},
			},
			want: want{
				res: []PageResponse{
					{
						ID:         "page-1",
						Title:      "t1",
						InviteCode: "invite-1",
						Links:      nil,
					},
					{
						ID:         "page-2",
						Title:      "t2",
						InviteCode: "invite-2",
						Links: []LinkResponse{
							{URL: "url1", Memo: "memo1", Priority: 1},
						},
					},
				},
				err: nil,
			},
		},
		{
			name: "success_list_by_invited_user",
			setup: func(m *mocks) {
				m.listUsecase.EXPECT().List(gomock.Any()).Return([]*dpage.Page{p3}, nil)
			},
			args: args{
				ctx: ctxuser.WithUser(context.Background(), invitedUser),
				req: ListRequest{},
			},
			want: want{
				res: []PageResponse{
					{
						ID:         "page-3",
						Title:      "t3",
						InviteCode: "invite-3",
						Links:      nil,
					},
				},
				err: nil,
			},
		},
		{
			name: "success_no_pages_found",
			setup: func(m *mocks) {
				m.listUsecase.EXPECT().List(gomock.Any()).Return(nil, nil)
			},
			args: args{
				ctx: ctxuser.WithUser(context.Background(), otherUser),
				req: ListRequest{},
			},
			want: want{
				res: nil,
				err: nil,
			},
		},
		{
			name: "usecase_error",
			setup: func(m *mocks) {
				m.listUsecase.EXPECT().List(gomock.Any()).Return(nil, errors.New("list error"))
			},
			args: args{
				ctx: ctxuser.WithUser(context.Background(), creator),
				req: ListRequest{},
			},
			want: want{
				res: nil,
				err: errors.New("list error"),
			},
		},
		{
			name: "user_not_found",
			setup: func(m *mocks) {
				m.listUsecase.EXPECT().List(gomock.Any()).Return(nil, duser.ErrUserNotFound)
			},
			args: args{
				ctx: ctxuser.WithUser(context.Background(), creator),
				req: ListRequest{},
			},
			want: want{
				res: nil,
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
				listUsecase: mocklist.NewMockListUsecase(ctrl),
			}
			if tt.setup != nil {
				tt.setup(m)
			}
			s := NewListService(m.listUsecase)
			got, err := s.List(tt.args.ctx, tt.args.req)
			if diff := cmp.Diff(tt.want.res, got); diff != "" {
				t.Errorf("response mismatch (-want +got):\n%s", diff)
			}
			testutil.EqualErr(t, tt.want.err, err)
		})
	}
}
