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

func TestListUsecase_List(t *testing.T) {
	type mocks struct {
		pageRepo *mockpage.MockPageRepository
	}
	type args struct {
		ctx     context.Context
		options []dpage.SearchOption
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
		setup func(m *mocks, t *testing.T)
		args  args
		want  want
	}{
		{
			name: "success_list_by_creator_user",
			setup: func(f *mocks, t *testing.T) {
				f.pageRepo.EXPECT().List(gomock.Any()).DoAndReturn(
					func(_ context.Context, opts ...dpage.SearchOption) ([]*dpage.Page, error) {
						if len(opts) != 0 {
							t.Fatalf("expected no options, got %d", len(opts))
						}
						return []*dpage.Page{p1, p2, p4}, nil
					},
				)
			},
			args: args{
				ctx:     ctxuser.WithUser(context.Background(), creator),
				options: nil,
			},
			want: want{
				pages: []*dpage.Page{p1, p2, p4},
				err:   nil,
			},
		},
		{
			name: "success_list_by_invited_user",
			setup: func(f *mocks, t *testing.T) {
				f.pageRepo.EXPECT().List(gomock.Any()).DoAndReturn(
					func(_ context.Context, opts ...dpage.SearchOption) ([]*dpage.Page, error) {
						if len(opts) != 0 {
							t.Fatalf("expected no options, got %d", len(opts))
						}
						return []*dpage.Page{p3}, nil
					},
				)
			},
			args: args{
				ctx:     ctxuser.WithUser(context.Background(), invitedUser),
				options: nil,
			},
			want: want{
				pages: []*dpage.Page{p3},
				err:   nil,
			},
		},
		{
			name: "success_no_pages_found",
			setup: func(f *mocks, t *testing.T) {
				f.pageRepo.EXPECT().List(gomock.Any()).DoAndReturn(
					func(_ context.Context, opts ...dpage.SearchOption) ([]*dpage.Page, error) {
						if len(opts) != 0 {
							t.Fatalf("expected no options, got %d", len(opts))
						}
						return nil, nil
					},
				)
			},
			args: args{
				ctx:     ctxuser.WithUser(context.Background(), otherUser),
				options: nil,
			},
			want: want{
				pages: nil,
				err:   nil,
			},
		},
		{
			name: "repo_error",
			setup: func(f *mocks, t *testing.T) {
				f.pageRepo.EXPECT().List(gomock.Any()).DoAndReturn(
					func(_ context.Context, opts ...dpage.SearchOption) ([]*dpage.Page, error) {
						if len(opts) != 0 {
							t.Fatalf("expected no options, got %d", len(opts))
						}
						return nil, errors.New("list error")
					},
				)
			},
			args: args{
				ctx:     ctxuser.WithUser(context.Background(), creator),
				options: nil,
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
				ctx:     context.Background(),
				options: nil,
			},
			want: want{
				pages: nil,
				err:   duser.ErrUserNotFound,
			},
		},
		{
			name: "success_with_ids_option",
			setup: func(f *mocks, t *testing.T) {
				f.pageRepo.EXPECT().List(gomock.Any(), gomock.Any()).DoAndReturn(
					func(_ context.Context, opts ...dpage.SearchOption) ([]*dpage.Page, error) {
						// Apply options to verify content
						var p dpage.SearchParams
						for _, o := range opts {
							o.Apply(&p)
						}
						wantIDs := []string{"page-1", "page-4"}
						if diff := cmp.Diff(wantIDs, p.IDs); diff != "" {
							t.Fatalf("SearchParams.IDs mismatch (-want +got):\n%s", diff)
						}
						if p.CreatedByUserID != "" {
							t.Fatalf("unexpected CreatedByUserID: %s", p.CreatedByUserID)
						}
						return []*dpage.Page{p1, p4}, nil
					},
				)
			},
			args: args{
				ctx:     ctxuser.WithUser(context.Background(), creator),
				options: []dpage.SearchOption{dpage.WithIDs([]string{"page-1", "page-4"})},
			},
			want: want{
				pages: []*dpage.Page{p1, p4},
				err:   nil,
			},
		},
		{
			name: "success_with_created_by_user_id_option",
			setup: func(f *mocks, t *testing.T) {
				f.pageRepo.EXPECT().List(gomock.Any(), gomock.Any()).DoAndReturn(
					func(_ context.Context, opts ...dpage.SearchOption) ([]*dpage.Page, error) {
						var p dpage.SearchParams
						for _, o := range opts {
							o.Apply(&p)
						}
						if p.CreatedByUserID != creator.ID() {
							t.Fatalf("CreatedByUserID want %s, got %s", creator.ID(), p.CreatedByUserID)
						}
						return []*dpage.Page{p1, p2}, nil
					},
				)
			},
			args: args{
				ctx:     ctxuser.WithUser(context.Background(), invitedUser),
				options: []dpage.SearchOption{dpage.WithCreatedByUserID(creator.ID())},
			},
			want: want{
				pages: []*dpage.Page{p1, p2},
				err:   nil,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			f := &mocks{pageRepo: mockpage.NewMockPageRepository(ctrl)}
			if tt.setup != nil {
				tt.setup(f, t)
			}
			u := NewListUsecase(f.pageRepo)
			got, err := u.List(tt.args.ctx, tt.args.options...)
			testutil.EqualErr(t, tt.want.err, err)
			if diff := cmp.Diff(tt.want.pages, got, cmp.AllowUnexported(dpage.Page{}, duser.User{})); diff != "" {
				t.Errorf("List() page mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
