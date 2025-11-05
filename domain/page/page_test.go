package page

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	di "github.com/naka-sei/tsudzuri/domain/user"
)

func TestPage_NewPage(t *testing.T) {
	type args struct {
		title      string
		createdBy  di.User
		inviteCode string
	}
	type want struct {
		title      string
		inviteCode string
		links      Links
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "new_empty_page",
			args: args{
				title:      "Title",
				createdBy:  di.User{},
				inviteCode: "invite123",
			},
			want: want{
				title:      "Title",
				inviteCode: "invite123",
				links:      Links{},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := NewPage(tt.args.title, tt.args.createdBy, tt.args.inviteCode)

			if diff := cmp.Diff(tt.want.title, got.Title(), cmp.AllowUnexported(Link{}, Page{})); diff != "" {
				t.Fatalf("title mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.want.inviteCode, got.InviteCode(), cmp.AllowUnexported(Link{}, Page{})); diff != "" {
				t.Fatalf("inviteCode mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.want.links, got.Links(), cmp.AllowUnexported(Link{}, Page{})); diff != "" {
				t.Fatalf("links mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestPage_Edit(t *testing.T) {
	type fields struct {
		page *Page
	}
	type args struct {
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
				},
			},
			args: args{
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
				},
			},
			args: args{
				title: "New",
				links: Links{
					{url: "a", memo: "A", priority: 1},
				},
			},
			want: want{
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
				},
			},
			args: args{
				title: "New",
				links: Links{
					{url: "no", memo: "Not Found", priority: 1},
				},
			},
			want: want{
				err: ErrNotFoundLink("no"),
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.fields.page.Edit(tt.args.title, tt.args.links)
			if tt.want.err != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.want.err)
				}
				if err.Error() != tt.want.err.Error() {
					t.Fatalf("error mismatch: want %v got %v", tt.want.err, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if diff := cmp.Diff(tt.want.page, tt.fields.page, cmp.AllowUnexported(Link{}, Page{}, di.User{})); diff != "" {
				t.Fatalf("page mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
