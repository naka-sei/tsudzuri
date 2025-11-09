package user

import (
	"testing"

	"github.com/google/go-cmp/cmp"
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
	email := "e@r.example"
	u := ReconstructUser("42", "u42", "google", &email)
	want := &User{id: "42", uid: "u42", provider: Provider("google"), email: &email}
	if diff := cmp.Diff(want, u, cmp.AllowUnexported(User{})); diff != "" {
		t.Fatalf("ReconstructUser mismatch (-want +got):\n%s", diff)
	}
}
