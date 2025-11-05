package page

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestLinks_AddLink(t *testing.T) {
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

			tt.fields.links.AddLink(tt.args.url, tt.args.memo)

			if diff := cmp.Diff(tt.want.links, tt.fields.links, cmp.AllowUnexported(Link{}, Page{})); diff != "" {
				t.Fatalf("links mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestLinks_RemoveLink(t *testing.T) {
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
					{url: "a", memo: "A", priority: 0},
					{url: "b", memo: "B", priority: 1},
				},
			},
			args: args{
				url: "a",
			},
			want: want{
				links: Links{
					{url: "b", memo: "B", priority: 0},
				},
			},
		},
		{
			name: "remove_not_found",
			fields: fields{
				links: Links{
					{url: "a", memo: "A", priority: 0},
				},
			},
			args: args{
				url: "no",
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

			err := tt.fields.links.RemoveLink(tt.args.url)
			if tt.want.err != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.want.err)
				}
				// compare by error string (NotFoundLink includes the url)
				if err.Error() != tt.want.err.Error() {
					t.Fatalf("error mismatch: want %v got %v", tt.want.err, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if diff := cmp.Diff(tt.want.links, tt.fields.links, cmp.AllowUnexported(Link{}, Page{})); diff != "" {
				t.Fatalf("links mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestLinks_editLink(t *testing.T) {
	type fields struct {
		links Links
	}
	type args struct {
		edit Link
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
			name: "edit_move_forward_and_memo",
			fields: fields{
				links: Links{
					{url: "b", memo: "B", priority: 1},
					{url: "a", memo: "A", priority: 3},
					{url: "c", memo: "C", priority: 2},
					{url: "d", memo: "D", priority: 4},
				},
			},
			args: args{
				edit: Link{url: "a", memo: "A-new", priority: 1},
			},
			want: want{
				links: Links{
					{url: "a", memo: "A-new", priority: 1},
					{url: "b", memo: "B", priority: 2},
					{url: "c", memo: "C", priority: 3},
					{url: "d", memo: "D", priority: 4},
				},
			},
		},
		{
			name: "edit_move_backward_and_memo",
			fields: fields{
				links: Links{
					{url: "b", memo: "B", priority: 1},
					{url: "a", memo: "A", priority: 3},
					{url: "c", memo: "C", priority: 2},
					{url: "d", memo: "D", priority: 4},
				},
			},
			args: args{
				edit: Link{url: "c", memo: "C-new", priority: 4},
			},
			want: want{
				links: Links{
					{url: "b", memo: "B", priority: 1},
					{url: "a", memo: "A", priority: 2},
					{url: "d", memo: "D", priority: 3},
					{url: "c", memo: "C-new", priority: 4},
				},
			},
		},
		{
			name: "edit_no_change",
			fields: fields{
				links: Links{
					{url: "x", memo: "X", priority: 1},
					{url: "y", memo: "Y", priority: 2},
				},
			},
			args: args{
				edit: Link{url: "x", memo: "X", priority: 1},
			},
			want: want{
				links: Links{
					{url: "x", memo: "X", priority: 1},
					{url: "y", memo: "Y", priority: 2},
				},
			},
		},
		{
			name: "edit_not_found",
			fields: fields{
				links: Links{
					{url: "x", memo: "X", priority: 0},
				},
			},
			args: args{
				edit: Link{url: "no", memo: "N", priority: 0},
			},
			want: want{err: ErrNotFoundLink("no")},
		},
		{
			name: "no_fields",
			fields: fields{
				links: Links{},
			},
			args: args{
				edit: Link{url: "x", memo: "X", priority: 1},
			},
			want: want{
				err: ErrNotFoundLink("x"),
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.fields.links.editLink(tt.args.edit)
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

			if diff := cmp.Diff(tt.want.links, tt.fields.links, cmp.AllowUnexported(Link{}, Page{})); diff != "" {
				t.Fatalf("links mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
