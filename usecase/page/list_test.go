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
	"github.com/naka-sei/tsudzuri/pkg/cmperr"
	ctxuser "github.com/naka-sei/tsudzuri/pkg/ctx/user"
)

func TestListUsecase_List(t *testing.T) {
	type mocks struct {
		pageRepo *mockpage.MockPageRepository
	}
	type args struct {
		ctx context.Context
	}
	type want struct {
		pages []*dpage.Page
		err   error
	}

	creator := duser.ReconstructUser("user-id-1", "uid-1", "anonymous", nil)
	invitedUser := duser.ReconstructUser("user-id-2", "uid-2", "invited", nil)
	otherUser := duser.ReconstructUser("user-id-3", "uid-3", "other", nil)

	p1 := dpage.ReconstructPage("page-1", "t1", *creator, "invite-1", dpage.Links{}, duser.Users{})
	p2 := dpage.ReconstructPage("page-2", "t2", *creator, "invite-2", dpage.Links{}, duser.Users{})
	p3 := dpage.ReconstructPage("page-3", "t3", *otherUser, "invite-3", dpage.Links{}, duser.Users{invitedUser})
	p4 := dpage.ReconstructPage("page-4", "t4", *otherUser, "invite-4", dpage.Links{}, duser.Users{creator})

	tests := []struct {
		name  string
		setup func(m *mocks)
		args  args
		want  want
	}{
		{
			name: "success_list_by_creator_user",
			setup: func(f *mocks) {
				f.pageRepo.EXPECT().List(gomock.Any(), "user-id-1", gomock.Any()).Return([]*dpage.Page{p1, p2, p4}, nil)
			},
			args: args{
				ctx: ctxuser.WithUser(context.Background(), creator),
			},
			want: want{
				pages: []*dpage.Page{p1, p2, p4},
				err:   nil,
			},
		},
		{
			name: "success_list_by_invited_user",
			setup: func(f *mocks) {
				f.pageRepo.EXPECT().List(gomock.Any(), "user-id-2", gomock.Any()).Return([]*dpage.Page{p3}, nil)
			},
			args: args{
				ctx: ctxuser.WithUser(context.Background(), invitedUser),
			},
			want: want{
				pages: []*dpage.Page{p3},
				err:   nil,
			},
		},
		{
			name: "success_no_pages_found",
			setup: func(f *mocks) {
				f.pageRepo.EXPECT().List(gomock.Any(), "user-id-3", gomock.Any()).Return(nil, nil)
			},
			args: args{
				ctx: ctxuser.WithUser(context.Background(), otherUser),
			},
			want: want{
				pages: nil,
				err:   nil,
			},
		},
		{
			name: "repo_error",
			setup: func(f *mocks) {
				f.pageRepo.EXPECT().List(gomock.Any(), "user-id-1", gomock.Any()).Return(nil, errors.New("list error"))
			},
			args: args{
				ctx: ctxuser.WithUser(context.Background(), creator),
			},
			want: want{
				pages: nil,
				err:   errors.New("list error"),
			},
		},
		{
			name:  "user_not_found",
			setup: nil,
			args: args{
				ctx: context.Background(),
			},
			want: want{
				pages: nil,
				err:   duser.ErrUserNotFound,
			},
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f := &mocks{
				pageRepo: mockpage.NewMockPageRepository(ctrl),
			}
			if tt.setup != nil {
				tt.setup(f)
			}
			u := NewListUsecase(f.pageRepo)
			got, err := u.List(tt.args.ctx)
			cmperr.Diff(t, tt.want.err, err)
			if diff := cmp.Diff(tt.want.pages, got, cmp.AllowUnexported(dpage.Page{}, duser.User{})); diff != "" {
				t.Errorf("List() page mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
