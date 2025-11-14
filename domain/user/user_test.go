package user

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/naka-sei/tsudzuri/pkg/testutil"
)

func TestUser_Login(t *testing.T) {
	type fields struct {
		user *User
	}
	type args struct {
		provider string
		email    *string
	}
	type want struct {
		err  error
		user *User
	}

	email := "e@example.com"

	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name:   "login_success",
			fields: fields{user: NewUser("u1")},
			args:   args{provider: "google", email: &email},
			want:   want{err: nil, user: &User{uid: "u1", provider: "google", email: &email}},
		},
		{
			name:   "login_no_email",
			fields: fields{user: NewUser("u2")},
			args:   args{provider: "google", email: nil},
			want:   want{err: ErrNoSpecifiedEmail, user: NewUser("u2")},
		},
		{
			name:   "login_invalid_provider",
			fields: fields{user: NewUser("u4")},
			args:   args{provider: "x", email: &email},
			want:   want{err: ErrInvalidProvider("x"), user: NewUser("u4")},
		},
		{
			name:   "login_already_logged_in",
			fields: fields{user: ReconstructUser("42", "u5", "google", &email)},
			args:   args{provider: "google", email: &email},
			want:   want{err: ErrAlreadyLoggedIn("google"), user: ReconstructUser("42", "u5", "google", &email)},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			u := tt.fields.user
			err := u.Login(tt.args.provider, tt.args.email)
			testutil.EqualErr(t, tt.want.err, err)

			if diff := cmp.Diff(tt.want.user, u, cmp.AllowUnexported(User{})); diff != "" {
				t.Fatalf("user mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestUser_NewUser(t *testing.T) {
	u := NewUser("uid123")
	want := &User{uid: "uid123", provider: ProviderAnonymous}
	if diff := cmp.Diff(want, u, cmp.AllowUnexported(User{})); diff != "" {
		t.Fatalf("NewUser mismatch (-want +got):\n%s", diff)
	}
}

func TestUser_ReconstructUser(t *testing.T) {
	type args struct {
		id       string
		uid      string
		provider string
		email    *string
		options  []ReconstructOption
	}
	type want struct {
		user *User
	}

	email := "e@r.example"

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "with_joined_pages",
			args: args{
				id:       "42",
				uid:      "u42",
				provider: "google",
				email:    &email,
				options:  []ReconstructOption{WithJoinedPageIDs([]string{"p1"})},
			},
			want: want{
				user: &User{id: "42", uid: "u42", provider: Provider("google"), email: &email, joinedPageIDs: []string{"p1"}},
			},
		},
	}

	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			u := ReconstructUser(tc.args.id, tc.args.uid, tc.args.provider, tc.args.email, tc.args.options...)
			if diff := cmp.Diff(tc.want.user, u, cmp.AllowUnexported(User{})); diff != "" {
				t.Fatalf("ReconstructUser mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestUser_JoinedPageIDs(t *testing.T) {
	type fields struct {
		user  *User
		input []string
	}
	type args struct {
		mutateReturned bool
		mutateInput    bool
	}
	type want struct {
		joinedIDs []string
	}

	tests := []struct {
		name  string
		setup func() fields
		args  args
		want  want
	}{
		{
			name: "returns_copy_of_joined_ids",
			setup: func() fields {
				ids := []string{"p1", "p2"}
				return fields{
					user:  ReconstructUser("42", "u42", "google", nil, WithJoinedPageIDs(ids)),
					input: ids,
				}
			},
			args: args{mutateReturned: true, mutateInput: true},
			want: want{joinedIDs: []string{"p1", "p2"}},
		},
		{
			name: "handles_nil_slice",
			setup: func() fields {
				return fields{user: ReconstructUser("43", "u43", "google", nil, WithJoinedPageIDs(nil))}
			},
			args: args{mutateReturned: false, mutateInput: false},
			want: want{joinedIDs: []string{}},
		},
	}

	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			f := tc.setup()
			expected := append([]string(nil), tc.want.joinedIDs...)

			if tc.args.mutateInput && len(f.input) > 0 {
				f.input[0] = "mutated-input"
			}

			returned := f.user.JoinedPageIDs()
			if diff := cmp.Diff(expected, returned, cmpopts.EquateEmpty()); diff != "" {
				t.Fatalf("JoinedPageIDs mismatch (-want +got):\n%s", diff)
			}

			if tc.args.mutateReturned && len(returned) > 0 {
				returned[0] = "mutated-returned"
			}

			again := f.user.JoinedPageIDs()
			if diff := cmp.Diff(expected, again, cmpopts.EquateEmpty()); diff != "" {
				t.Fatalf("JoinedPageIDs should remain unchanged (-want +got):\n%s", diff)
			}
		})
	}
}
