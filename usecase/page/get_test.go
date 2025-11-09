package page

import (
	"context"
	"errors"
	"testing"

	cmp "github.com/google/go-cmp/cmp"
	"go.uber.org/mock/gomock"

	dpage "github.com/naka-sei/tsudzuri/domain/page"
	mockpage "github.com/naka-sei/tsudzuri/domain/page/mock/mock_page"
	duser "github.com/naka-sei/tsudzuri/domain/user"
	ctxuser "github.com/naka-sei/tsudzuri/pkg/ctx/user"
	"github.com/naka-sei/tsudzuri/pkg/testutil"
)

func TestGetUsecase_Get(t *testing.T) {
	type mocks struct {
		pageRepo *mockpage.MockPageRepository
	}
	type args struct {
		ctx    context.Context
		pageID string
	}
	type want struct {
		page *dpage.Page
		err  error
	}

	user := duser.ReconstructUser("user-id-1", "uid-1", "anonymous", nil)
	other := duser.ReconstructUser("user-id-2", "uid-2", "anonymous", nil)
	invitedUser := duser.ReconstructUser("user-id-3", "uid-3", "invited", nil)

	p1 := dpage.ReconstructPage("page-1", "t1", *user, "invite", dpage.Links{}, duser.Users{})
	p2 := dpage.ReconstructPage("page-2", "t2", *other, "invite", dpage.Links{}, duser.Users{})
	p3 := dpage.ReconstructPage("page-3", "t3", *other, "invite", dpage.Links{}, duser.Users{invitedUser})

	tests := []struct {
		name  string
		setup func(m *mocks)
		args  args
		want  want
	}{
		{
			name: "success_by_creator",
			setup: func(m *mocks) {
				m.pageRepo.EXPECT().Get(gomock.Any(), "page-1").Return(p1, nil)
			},
			args: args{
				ctx:    ctxuser.WithUser(context.Background(), user),
				pageID: "page-1",
			},
			want: want{
				page: p1,
				err:  nil,
			},
		},
		{
			name: "success_by_invited_user",
			setup: func(m *mocks) {
				m.pageRepo.EXPECT().Get(gomock.Any(), "page-3").Return(p3, nil)
			},
			args: args{
				ctx:    ctxuser.WithUser(context.Background(), invitedUser),
				pageID: "page-3",
			},
			want: want{
				page: p3,
				err:  nil,
			},
		},
		{
			name: "repo_error",
			setup: func(m *mocks) {
				m.pageRepo.EXPECT().Get(gomock.Any(), "page-1").Return(nil, errors.New("get error"))
			},
			args: args{
				ctx:    ctxuser.WithUser(context.Background(), user),
				pageID: "page-1",
			},
			want: want{
				page: nil,
				err:  errors.New("get error"),
			},
		},
		{
			name: "page_not_found",
			setup: func(m *mocks) {
				m.pageRepo.EXPECT().Get(gomock.Any(), "page-1").Return(nil, nil)
			},
			args: args{
				ctx:    ctxuser.WithUser(context.Background(), user),
				pageID: "page-1",
			},
			want: want{
				page: nil,
				err:  ErrPageNotFound,
			},
		},
		{
			name: "user_not_found",
			setup: func(m *mocks) {
				m.pageRepo.EXPECT().Get(gomock.Any(), "page-1").Return(p1, nil)
			},
			args: args{
				ctx:    context.Background(),
				pageID: "page-1",
			},
			want: want{
				page: nil,
				err:  duser.ErrUserNotFound,
			},
		},
		{
			name: "unauthorized",
			setup: func(m *mocks) {
				m.pageRepo.EXPECT().Get(gomock.Any(), "page-2").Return(p2, nil)
			},
			args: args{
				ctx:    ctxuser.WithUser(context.Background(), user),
				pageID: "page-2",
			},
			want: want{
				page: nil,
				err:  dpage.ErrNotCreatedByUser,
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
				pageRepo: mockpage.NewMockPageRepository(ctrl),
			}
			if tt.setup != nil {
				tt.setup(m)
			}
			u := NewGetUsecase(m.pageRepo)
			got, err := u.Get(tt.args.ctx, tt.args.pageID)
			if diff := cmp.Diff(tt.want.page, got, cmp.AllowUnexported(dpage.Page{}, duser.User{})); diff != "" {
				t.Errorf("page mismatch (-want +got):\n%s", diff)
			}
			testutil.EqualErr(t, tt.want.err, err)
		})
	}
}
