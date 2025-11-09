package page

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	di "github.com/naka-sei/tsudzuri/domain/user"
	"github.com/naka-sei/tsudzuri/pkg/testutil"
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
			name: "no_title_provided",
			args: args{
				title:     "",
				createdBy: &di.User{},
			},
			want: want{
				err: ErrNoTitleProvided,
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
			testutil.EqualErr(t, tt.want.err, err)

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
		{
			name: "nil_user",
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
				user:  nil,
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
					},
					invitedUsers: di.Users{&di.User{}},
				},
				err: ErrNoUserProvided,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.fields.page.Edit(tt.args.user, tt.args.title, tt.args.links)
			testutil.EqualErr(t, tt.want.err, err)

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
			testutil.EqualErr(t, tt.want.err, err)
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
			testutil.EqualErr(t, tt.want.err, err)
		})
	}
}

func TestPage_AddLink(t *testing.T) {
	type fields struct {
		page *Page
	}
	type args struct {
		user *di.User
		url  string
		memo string
	}
	type want struct {
		page *Page
		err  error
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
			name: "success_by_creator",
			fields: fields{
				page: &Page{
					title:      "Title",
					createdBy:  *creator,
					inviteCode: "code",
					links:      Links{},
				},
			},
			args: args{
				user: creator,
				url:  "https://example.com",
				memo: "Example",
			},
			want: want{
				page: &Page{
					title:      "Title",
					createdBy:  *creator,
					inviteCode: "code",
					links: Links{
						{url: "https://example.com", memo: "Example", priority: 0},
					},
				},
			},
		},
		{
			name: "success_by_invited_user",
			fields: fields{
				page: &Page{
					title:        "Title",
					createdBy:    *creator,
					inviteCode:   "code",
					links:        Links{},
					invitedUsers: di.Users{invited},
				},
			},
			args: args{
				user: invited,
				url:  "https://example.com",
				memo: "Example",
			},
			want: want{
				page: &Page{
					title:      "Title",
					createdBy:  *creator,
					inviteCode: "code",
					links: Links{
						{url: "https://example.com", memo: "Example", priority: 0},
					},
					invitedUsers: di.Users{invited},
				},
			},
		},
		{
			name: "unauthorized_nil_user",
			fields: fields{
				page: &Page{
					title:      "Title",
					createdBy:  *creator,
					inviteCode: "code",
					links:      Links{},
				},
			},
			args: args{
				user: nil,
				url:  "https://example.com",
				memo: "Example",
			},
			want: want{
				page: &Page{
					title:      "Title",
					createdBy:  *creator,
					inviteCode: "code",
					links:      Links{},
				},
				err: ErrNoUserProvided,
			},
		},
		{
			name: "unauthorized_not_creator",
			fields: fields{
				page: &Page{
					title:      "Title",
					createdBy:  *creator,
					inviteCode: "code",
					links:      Links{},
				},
			},
			args: args{
				user: other,
				url:  "https://example.com",
				memo: "Example",
			},
			want: want{
				page: &Page{
					title:      "Title",
					createdBy:  *creator,
					inviteCode: "code",
					links:      Links{},
				},
				err: ErrNotCreatedByUser,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.fields.page.AddLink(tt.args.user, tt.args.url, tt.args.memo)
			testutil.EqualErr(t, tt.want.err, err)

			if diff := cmp.Diff(tt.want.page, tt.fields.page, cmp.AllowUnexported(Link{}, Page{}, di.User{})); diff != "" {
				t.Fatalf("page mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestPage_RemoveLink(t *testing.T) {
	type fields struct {
		page *Page
	}
	type args struct {
		user *di.User
		url  string
	}
	type want struct {
		page *Page
		err  error
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
			name: "success_by_creator",
			fields: fields{
				page: &Page{
					title:      "Title",
					createdBy:  *creator,
					inviteCode: "code",
					links: Links{
						{url: "https://a.com", memo: "A", priority: 1},
						{url: "https://b.com", memo: "B", priority: 2},
					},
				},
			},
			args: args{
				user: creator,
				url:  "https://a.com",
			},
			want: want{
				page: &Page{
					title:      "Title",
					createdBy:  *creator,
					inviteCode: "code",
					links: Links{
						{url: "https://b.com", memo: "B", priority: 1},
					},
				},
			},
		},
		{
			name: "success_by_invited_user",
			fields: fields{
				page: &Page{
					title:      "Title",
					createdBy:  *creator,
					inviteCode: "code",
					links: Links{
						{url: "https://a.com", memo: "A", priority: 1},
					},
					invitedUsers: di.Users{invited},
				},
			},
			args: args{
				user: invited,
				url:  "https://a.com",
			},
			want: want{
				page: &Page{
					title:        "Title",
					createdBy:    *creator,
					inviteCode:   "code",
					links:        Links{},
					invitedUsers: di.Users{invited},
				},
			},
		},
		{
			name: "link_not_found",
			fields: fields{
				page: &Page{
					title:      "Title",
					createdBy:  *creator,
					inviteCode: "code",
					links: Links{
						{url: "https://a.com", memo: "A", priority: 1},
					},
				},
			},
			args: args{
				user: creator,
				url:  "https://notfound.com",
			},
			want: want{
				page: &Page{
					title:      "Title",
					createdBy:  *creator,
					inviteCode: "code",
					links: Links{
						{url: "https://a.com", memo: "A", priority: 1},
					},
				},
				err: ErrNotFoundLink("https://notfound.com"),
			},
		},
		{
			name: "unauthorized_nil_user",
			fields: fields{
				page: &Page{
					title:      "Title",
					createdBy:  *creator,
					inviteCode: "code",
					links: Links{
						{url: "https://a.com", memo: "A", priority: 1},
					},
				},
			},
			args: args{
				user: nil,
				url:  "https://a.com",
			},
			want: want{
				page: &Page{
					title:      "Title",
					createdBy:  *creator,
					inviteCode: "code",
					links: Links{
						{url: "https://a.com", memo: "A", priority: 1},
					},
				},
				err: ErrNoUserProvided,
			},
		},
		{
			name: "unauthorized_not_creator",
			fields: fields{
				page: &Page{
					title:      "Title",
					createdBy:  *creator,
					inviteCode: "code",
					links: Links{
						{url: "https://a.com", memo: "A", priority: 1},
					},
				},
			},
			args: args{
				user: other,
				url:  "https://a.com",
			},
			want: want{
				page: &Page{
					title:      "Title",
					createdBy:  *creator,
					inviteCode: "code",
					links: Links{
						{url: "https://a.com", memo: "A", priority: 1},
					},
				},
				err: ErrNotCreatedByUser,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.fields.page.RemoveLink(tt.args.user, tt.args.url)
			testutil.EqualErr(t, tt.want.err, err)

			if diff := cmp.Diff(tt.want.page, tt.fields.page, cmp.AllowUnexported(Link{}, Page{}, di.User{})); diff != "" {
				t.Fatalf("page mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestReconstructPage(t *testing.T) {
	type args struct {
		id           string
		title        string
		createdBy    di.User
		inviteCode   string
		links        Links
		invitedUsers di.Users
	}
	tests := []struct {
		name string
		args args
		want *Page
	}{
		{
			name: "basic",
			args: args{
				id:         "page-id",
				title:      "Title",
				createdBy:  di.User{},
				inviteCode: "code",
				links: Links{
					{url: "https://a.com", memo: "A", priority: 1},
				},
				invitedUsers: di.Users{&di.User{}},
			},
			want: &Page{
				id:         "page-id",
				title:      "Title",
				createdBy:  di.User{},
				inviteCode: "code",
				links: Links{
					{url: "https://a.com", memo: "A", priority: 1},
				},
				invitedUsers: di.Users{&di.User{}},
			},
		},
		{
			name: "empty",
			args: args{
				id:           "",
				title:        "",
				createdBy:    di.User{},
				inviteCode:   "",
				links:        Links{},
				invitedUsers: di.Users{},
			},
			want: &Page{
				id:           "",
				title:        "",
				createdBy:    di.User{},
				inviteCode:   "",
				links:        Links{},
				invitedUsers: di.Users{},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ReconstructPage(tt.args.id, tt.args.title, tt.args.createdBy, tt.args.inviteCode, tt.args.links, tt.args.invitedUsers)
			if diff := cmp.Diff(tt.want, got, cmp.AllowUnexported(Link{}, Page{}, di.User{})); diff != "" {
				t.Fatalf("page mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
