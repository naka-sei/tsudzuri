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
	mockget "github.com/naka-sei/tsudzuri/usecase/page/mock/mock_get"
)

func TestGetService_Get(t *testing.T) {
	type mocks struct {
		getUsecase *mockget.MockGetUsecase
	}
	type args struct {
		ctx context.Context
		req GetRequest
	}
	type want struct {
		res PageResponse
		err error
	}

	user := duser.ReconstructUser("user-id-1", "uid-1", "anonymous", nil)
	invitedUser := duser.ReconstructUser("user-id-3", "uid-3", "invited", nil)

	p1 := dpage.ReconstructPage("page-1", "t1", *user, "invite", dpage.Links{}, duser.Users{})
	p3 := dpage.ReconstructPage("page-3", "t3", *user, "invite", dpage.Links{dpage.ReconstructLink("url1", "memo1", 1)}, duser.Users{invitedUser})

	tests := []struct {
		name  string
		setup func(m *mocks)
		args  args
		want  want
	}{
		{
			name: "success_by_creator",
			setup: func(m *mocks) {
				m.getUsecase.EXPECT().Get(gomock.Any(), "page-1").Return(p1, nil)
			},
			args: args{
				ctx: ctxuser.WithUser(context.Background(), user),
				req: GetRequest{PageID: "page-1"},
			},
			want: want{
				res: PageResponse{
					ID:         "page-1",
					Title:      "t1",
					InviteCode: "invite",
					Links:      nil,
				},
				err: nil,
			},
		},
		{
			name: "success_by_invited_user_with_links",
			setup: func(m *mocks) {
				m.getUsecase.EXPECT().Get(gomock.Any(), "page-3").Return(p3, nil)
			},
			args: args{
				ctx: ctxuser.WithUser(context.Background(), invitedUser),
				req: GetRequest{PageID: "page-3"},
			},
			want: want{
				res: PageResponse{
					ID:         "page-3",
					Title:      "t3",
					InviteCode: "invite",
					Links: []LinkResponse{
						{URL: "url1", Memo: "memo1", Priority: 1},
					},
				},
				err: nil,
			},
		},
		{
			name: "usecase_error",
			setup: func(m *mocks) {
				m.getUsecase.EXPECT().Get(gomock.Any(), "page-1").Return(nil, errors.New("get error"))
			},
			args: args{
				ctx: ctxuser.WithUser(context.Background(), user),
				req: GetRequest{PageID: "page-1"},
			},
			want: want{
				res: PageResponse{},
				err: errors.New("get error"),
			},
		},
		{
			name: "user_not_found",
			setup: func(m *mocks) {
				m.getUsecase.EXPECT().Get(gomock.Any(), "page-1").Return(nil, duser.ErrUserNotFound)
			},
			args: args{
				ctx: ctxuser.WithUser(context.Background(), user),
				req: GetRequest{PageID: "page-1"},
			},
			want: want{
				res: PageResponse{},
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
				getUsecase: mockget.NewMockGetUsecase(ctrl),
			}
			if tt.setup != nil {
				tt.setup(m)
			}
			s := NewGetService(m.getUsecase)
			got, err := s.Get(tt.args.ctx, tt.args.req)
			if diff := cmp.Diff(tt.want.res, got); diff != "" {
				t.Errorf("response mismatch (-want +got):\n%s", diff)
			}
			testutil.EqualErr(t, tt.want.err, err)
		})
	}
}
