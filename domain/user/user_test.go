package user

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestUser_Login(t *testing.T) {
	type fields struct {
		user *User
	}
	type args struct {
		uid      string
		provider Provider
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
			args:   args{uid: "u1", provider: ProviderGoogle, email: &email},
			want:   want{err: nil, user: &User{uid: "u1", provider: ProviderGoogle, email: &email}},
		},
		{
			name:   "login_no_email",
			fields: fields{user: NewUser("u2")},
			args:   args{uid: "u2", provider: ProviderGoogle, email: nil},
			want:   want{err: ErrNoSpecifiedEmail},
		},
		{
			name:   "login_invalid_uid",
			fields: fields{user: NewUser("u3")},
			args:   args{uid: "other", provider: ProviderGoogle, email: &email},
			want:   want{err: ErrInvalidUID("other")},
		},
		{
			name:   "login_invalid_provider",
			fields: fields{user: NewUser("u4")},
			args:   args{uid: "u4", provider: Provider("x"), email: &email},
			want:   want{err: ErrInvalidProvider(Provider("x"))},
		},
		{
			name:   "login_already_logged_in",
			fields: fields{user: ReconstructUser(1, "u5", "google", &email)},
			args:   args{uid: "u5", provider: ProviderGoogle, email: &email},
			want:   want{err: ErrAlreadyLoggedIn(ProviderGoogle)},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			u := tt.fields.user
			err := u.Login(tt.args.uid, tt.args.provider, tt.args.email)
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
	u := ReconstructUser(42, "u42", "google", &email)
	want := &User{id: 42, uid: "u42", provider: Provider("google"), email: &email}
	if diff := cmp.Diff(want, u, cmp.AllowUnexported(User{})); diff != "" {
		t.Fatalf("ReconstructUser mismatch (-want +got):\n%s", diff)
	}
}
