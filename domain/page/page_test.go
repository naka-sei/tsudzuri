package page

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	di "github.com/naka-sei/tsudzuri/domain/user"
	"github.com/naka-sei/tsudzuri/pkg/cmperr"
)

func TestPage_NewPage(t *testing.T) {
	type args struct {
		title     string
		createdBy *di.User
	}
	type want struct {
		page *Page
		err  error
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "new_empty_page",
			args: args{
				title:     "Title",
				createdBy: &di.User{},
			},
			want: want{
				page: &Page{
					title:      "Title",
					createdBy:  di.User{},
					inviteCode: "sample",
					links:      Links{},
				},
			},
		},
		{
			name: "no_user_provided",
			args: args{
				title:     "Title",
				createdBy: nil,
			},
			want: want{
				err: ErrNoUserProvided,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := NewPage(tt.args.title, tt.args.createdBy)
			cmperr.Diff(t, tt.want.err, err)

			if diff := cmp.Diff(tt.want.page, got, cmp.AllowUnexported(Link{}, Page{}, di.User{})); diff != "" {
				t.Fatalf("page mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestPage_Edit(t *testing.T) {
	type fields struct {
		page *Page
	}
	type args struct {
		user  *di.User
		title string
		links Links
	}
	type want struct {
		page *Page
		err  error
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name: "edit_title_and_reorder",
			fields: fields{
				page: &Page{
					title:      "Old",
					createdBy:  di.User{},
					inviteCode: "code",
					links: Links{
						{url: "a", memo: "A", priority: 1},
						{url: "b", memo: "B", priority: 2},
					},
					invitedUsers: di.Users{&di.User{}},
				},
			},
			args: args{
				user:  &di.User{},
				title: "New",
				links: Links{
					{url: "b", memo: "B-new", priority: 1},
					{url: "a", memo: "A", priority: 2},
				},
			},
			want: want{
				page: &Page{
					title:      "New",
					createdBy:  di.User{},
					inviteCode: "code",
					links: Links{
						{url: "b", memo: "B-new", priority: 1},
						{url: "a", memo: "A", priority: 2},
					},
					invitedUsers: di.Users{&di.User{}},
				},
			},
		},
		{
			name: "invalid_links_length",
			fields: fields{
				page: &Page{
					title:      "Old",
					createdBy:  di.User{},
					inviteCode: "code",
					links: Links{
						{url: "a", memo: "A", priority: 1},
						{url: "b", memo: "B", priority: 2},
					},
					invitedUsers: di.Users{&di.User{}},
				},
			},
			args: args{
				user:  &di.User{},
				title: "New",
				links: Links{
					{url: "a", memo: "A", priority: 1},
				},
			},
			want: want{
				page: &Page{
					title:      "Old",
					createdBy:  di.User{},
					inviteCode: "code",
					links: Links{
						{url: "a", memo: "A", priority: 1},
						{url: "b", memo: "B", priority: 2},
					},
					invitedUsers: di.Users{&di.User{}},
				},
				err: ErrInvalidLinksLength,
			},
		},
		{
			name: "invalid_links_length_extra",
			fields: fields{
				page: &Page{
					title:      "Old",
					createdBy:  di.User{},
					inviteCode: "code",
					links: Links{
						{url: "a", memo: "A", priority: 1},
					},
					invitedUsers: di.Users{&di.User{}},
				},
			},
			args: args{
				user:  &di.User{},
				title: "New",
				links: Links{
					{url: "a", memo: "A", priority: 1},
					{url: "b", memo: "B", priority: 2},
				},
			},
			want: want{
				page: &Page{
					title:      "Old",
					createdBy:  di.User{},
					inviteCode: "code",
					links: Links{
						{url: "a", memo: "A", priority: 1},
					},
					invitedUsers: di.Users{&di.User{}},
				},
				err: ErrInvalidLinksLength,
			},
		},
		{
			name: "not_found_link",
			fields: fields{
				page: &Page{
					title:      "Old",
					createdBy:  di.User{},
					inviteCode: "code",
					links: Links{
						{url: "a", memo: "A", priority: 1},
					},
					invitedUsers: di.Users{&di.User{}},
				},
			},
			args: args{
				user:  &di.User{},
				title: "New",
				links: Links{
					{url: "no", memo: "Not Found", priority: 1},
				},
			},
			want: want{
				page: &Page{
					title:      "Old",
					createdBy:  di.User{},
					inviteCode: "code",
					links: Links{
						{url: "a", memo: "A", priority: 1},
					},
					invitedUsers: di.Users{&di.User{}},
				},
				err: ErrNotFoundLink("no"),
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.fields.page.Edit(tt.args.user, tt.args.title, tt.args.links)
			cmperr.Diff(t, tt.want.err, err)

			if diff := cmp.Diff(tt.want.page, tt.fields.page, cmp.AllowUnexported(Link{}, Page{}, di.User{})); diff != "" {
				t.Fatalf("page mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestPage_Authorize(t *testing.T) {
	type fields struct {
		page *Page
	}
	type args struct {
		user *di.User
	}
	type want struct {
		err error
	}

	creator := di.ReconstructUser("creator-id", "uid-c", "anonymous", nil)
	invited := di.ReconstructUser("invited-id", "uid-i", "anonymous", nil)
	other := di.ReconstructUser("other-id", "uid-o", "anonymous", nil)

	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name: "invited_user_allowed",
			fields: fields{
				page: &Page{
					createdBy:    *creator,
					invitedUsers: di.Users{invited},
				},
			},
			args: args{user: invited},
			want: want{err: nil},
		},
		{
			name: "creator_allowed",
			fields: fields{
				page: &Page{
					createdBy:    *creator,
					invitedUsers: di.Users{},
				},
			},
			args: args{user: creator},
			want: want{err: nil},
		},
		{
			name: "no_user_provided",
			fields: fields{
				page: &Page{
					createdBy:    *creator,
					invitedUsers: di.Users{},
				},
			},
			args: args{user: nil},
			want: want{err: ErrNoUserProvided},
		},
		{
			name: "unauthorized",
			fields: fields{
				page: &Page{
					createdBy:    *creator,
					invitedUsers: di.Users{},
				},
			},
			args: args{user: other},
			want: want{err: ErrNotCreatedByUser},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.fields.page.Authorize(tt.args.user)
			cmperr.Diff(t, tt.want.err, err)
		})
	}
}

func TestPage_validateCreatedBy(t *testing.T) {
	type fields struct {
		page *Page
	}
	type args struct {
		user *di.User
	}
	type want struct {
		err error
	}

	creator := di.ReconstructUser("creator-id", "uid-c", "anonymous", nil)
	other := di.ReconstructUser("other-id", "uid-o", "anonymous", nil)

	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name:   "nil_user",
			fields: fields{page: &Page{createdBy: *creator}},
			args:   args{user: nil},
			want:   want{err: ErrNoUserProvided},
		},
		{
			name:   "valid_creator",
			fields: fields{page: &Page{createdBy: *creator}},
			args:   args{user: creator},
			want:   want{err: nil},
		},
		{
			name:   "invalid_creator",
			fields: fields{page: &Page{createdBy: *creator}},
			args:   args{user: other},
			want:   want{err: ErrNotCreatedByUser},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.fields.page.validateCreatedBy(tt.args.user)
			cmperr.Diff(t, tt.want.err, err)
		})
	}
}
