package page

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/naka-sei/tsudzuri/pkg/testutil"
)

func TestLinks_addLink(t *testing.T) {
	type fields struct {
		links Links
	}
	type args struct {
		url  string
		memo string
	}
	type want struct {
		links Links
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name: "add_to_empty",
			fields: fields{
				links: Links{},
			},
			args: args{
				url:  "https://x",
				memo: "m",
			},
			want: want{
				links: Links{
					{url: "https://x", memo: "m", priority: 0},
				},
			},
		},
		{
			name: "add_to_existing",
			fields: fields{
				Links{
					{url: "a", memo: "A", priority: 0},
				},
			},
			args: args{
				url:  "https://b",
				memo: "B",
			},
			want: want{
				links: Links{
					{url: "a", memo: "A", priority: 0},
					{url: "https://b", memo: "B", priority: 1},
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tt.fields.links.addLink(tt.args.url, tt.args.memo)

			if diff := cmp.Diff(tt.want.links, tt.fields.links, cmp.AllowUnexported(Link{}, Page{})); diff != "" {
				t.Fatalf("links mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestLinks_removeLink(t *testing.T) {
	type fields struct {
		links Links
	}
	type args struct {
		url string
	}
	type want struct {
		links Links
		err   error
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name: "remove_existing",
			fields: fields{
				links: Links{
					{url: "a", memo: "A", priority: 1},
					{url: "b", memo: "B", priority: 2},
				},
			},
			args: args{
				url: "a",
			},
			want: want{
				links: Links{
					{url: "b", memo: "B", priority: 1},
				},
			},
		},
		{
			name: "remove_not_found",
			fields: fields{
				links: Links{
					{url: "a", memo: "A", priority: 1},
				},
			},
			args: args{
				url: "no",
			},
			want: want{
				links: Links{
					{url: "a", memo: "A", priority: 1},
				},
				err: ErrNotFoundLink("no"),
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.fields.links.removeLink(tt.args.url)
			testutil.EqualErr(t, tt.want.err, err)

			if diff := cmp.Diff(tt.want.links, tt.fields.links, cmp.AllowUnexported(Link{}, Page{})); diff != "" {
				t.Fatalf("links mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestLinks_editLinks(t *testing.T) {
	type fields struct {
		links Links
	}
	type args struct {
		links Links
	}
	type want struct {
		links Links
		err   error
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name: "edit_success_reorder",
			fields: fields{
				links: Links{
					{url: "a", memo: "A", priority: 1},
					{url: "b", memo: "B", priority: 2},
					{url: "c", memo: "C", priority: 3},
				},
			},
			args: args{
				links: Links{
					{url: "b", memo: "B-mod", priority: 3},
					{url: "a", memo: "A-mod", priority: 2},
					{url: "c", memo: "C-mod", priority: 1},
				},
			},
			want: want{
				links: Links{
					{url: "c", memo: "C-mod", priority: 1},
					{url: "a", memo: "A-mod", priority: 2},
					{url: "b", memo: "B-mod", priority: 3},
				},
				err: nil,
			},
		},
		{
			name: "edit_not_found",
			fields: fields{
				links: Links{
					{url: "a", memo: "A", priority: 1},
				},
			},
			args: args{
				links: Links{
					{url: "no", memo: "X", priority: 1},
				},
			},
			want: want{
				links: Links{
					{url: "a", memo: "A", priority: 1},
				},
				err: ErrNotFoundLink("no"),
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.fields.links.editLinks(tt.args.links)
			testutil.EqualErr(t, tt.want.err, err)

			if diff := cmp.Diff(tt.want.links, tt.fields.links, cmp.AllowUnexported(Link{}, Page{})); diff != "" {
				t.Fatalf("links mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestReconstructLink(t *testing.T) {
	type args struct {
		url      string
		memo     string
		priority int
	}
	tests := []struct {
		name string
		args args
		want Link
	}{
		{
			name: "basic",
			args: args{url: "https://x", memo: "memo", priority: 3},
			want: Link{url: "https://x", memo: "memo", priority: 3},
		},
		{
			name: "empty_fields",
			args: args{url: "", memo: "", priority: 0},
			want: Link{url: "", memo: "", priority: 0},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ReconstructLink(tt.args.url, tt.args.memo, tt.args.priority)
			if diff := cmp.Diff(tt.want, got, cmp.AllowUnexported(Link{})); diff != "" {
				t.Fatalf("link mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
