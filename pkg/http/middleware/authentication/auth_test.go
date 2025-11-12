package authentication

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"firebase.google.com/go/v4/auth"
	"github.com/google/go-cmp/cmp"
	"go.uber.org/mock/gomock"

	duser "github.com/naka-sei/tsudzuri/domain/user"
	mockuser "github.com/naka-sei/tsudzuri/domain/user/mock/mock_user"
	"github.com/naka-sei/tsudzuri/pkg/cache"
	ctxuser "github.com/naka-sei/tsudzuri/pkg/ctx/user"
)

func TestAuthHTTPMiddleware(t *testing.T) {
	t.Parallel()

	type fields struct {
		userRepo *mockuser.MockUserRepository
		cache    cache.Cache[*duser.User]
	}

	type args struct {
		method     string
		target     string
		authHeader string
	}

	type want struct {
		user       *duser.User
		nextCalled bool
		authCalled bool
	}

	type verify func(t *testing.T, f *fields)

	cachedUser := duser.ReconstructUser("1", "cached-uid", string(duser.ProviderAnonymous), nil)
	fetchedUser := duser.ReconstructUser("2", "user-uid", string(duser.ProviderAnonymous), nil)

	tests := []struct {
		name   string
		uid    string
		setup  func(f *fields)
		args   args
		want   want
		verify verify
	}{
		{
			name: "cache_hit",
			uid:  cachedUser.UID(),
			setup: func(f *fields) {
				f.cache = cache.NewMemoryCache[*duser.User](time.Minute)
				f.cache.Set(context.Background(), cachedUser.UID(), cachedUser)
				f.userRepo.EXPECT().Get(gomock.Any(), gomock.Any()).Times(0)
			},
			args: args{
				method:     http.MethodGet,
				target:     "/",
				authHeader: "token",
			},
			want: want{
				user:       cachedUser,
				nextCalled: true,
				authCalled: true,
			},
		},
		{
			name: "cache_miss_fetches_and_stores",
			uid:  fetchedUser.UID(),
			setup: func(f *fields) {
				f.cache = cache.NewMemoryCache[*duser.User](time.Minute)
				f.userRepo.EXPECT().Get(gomock.Any(), fetchedUser.UID()).Return(fetchedUser, nil)
			},
			args: args{
				method:     http.MethodGet,
				target:     "/",
				authHeader: "token",
			},
			want: want{
				user:       fetchedUser,
				nextCalled: true,
				authCalled: true,
			},
			verify: func(t *testing.T, f *fields) {
				t.Helper()
				cached, ok := f.cache.Get(context.Background(), fetchedUser.UID())
				if !ok {
					t.Fatalf("expected user cached after fetch")
				}
				if diff := cmp.Diff(fetchedUser, cached, cmp.AllowUnexported(duser.User{})); diff != "" {
					t.Fatalf("cached user mismatch (-want +got):\n%s", diff)
				}
			},
		},
		{
			name: "public_route_bypasses_auth",
			setup: func(f *fields) {
				f.cache = cache.NewMemoryCache[*duser.User](time.Minute)
				f.userRepo.EXPECT().Get(gomock.Any(), gomock.Any()).Times(0)
			},
			args: args{
				method:     http.MethodPost,
				target:     "/api/v1/users",
				authHeader: "",
			},
			want: want{
				user:       nil,
				nextCalled: true,
				authCalled: false,
			},
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			f := &fields{
				userRepo: mockuser.NewMockUserRepository(ctrl),
				cache:    cache.NewMemoryCache[*duser.User](time.Minute),
			}
			if tt.setup != nil {
				tt.setup(f)
			}

			authenticator := &stubAuthenticator{token: &auth.Token{UID: tt.uid}}
			req := httptest.NewRequest(tt.args.method, tt.args.target, nil)
			if tt.args.authHeader != "" {
				req.Header.Set("Authorization", tt.args.authHeader)
			}
			rr := httptest.NewRecorder()

			nextCalled := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				if tt.want.user == nil {
					if _, ok := ctxuser.UserFromContext(r.Context()); ok {
						t.Fatalf("did not expect user in context")
					}
					return
				}

				got, ok := ctxuser.UserFromContext(r.Context())
				if !ok {
					t.Fatalf("expected user in context")
				}
				if diff := cmp.Diff(tt.want.user, got, cmp.AllowUnexported(duser.User{})); diff != "" {
					t.Fatalf("user mismatch (-want +got):\n%s", diff)
				}
			})

			handler := AuthHTTPMiddleware(authenticator, f.userRepo, f.cache)(next)
			handler.ServeHTTP(rr, req)

			if nextCalled != tt.want.nextCalled {
				t.Fatalf("next handler called = %t, want %t", nextCalled, tt.want.nextCalled)
			}
			if authenticator.called != tt.want.authCalled {
				t.Fatalf("authenticator called = %t, want %t", authenticator.called, tt.want.authCalled)
			}

			if tt.verify != nil {
				tt.verify(t, f)
			}
		})
	}
}

type stubAuthenticator struct {
	token  *auth.Token
	err    error
	called bool
}

func (s *stubAuthenticator) VerifyIDToken(ctx context.Context, idToken string) (*auth.Token, error) {
	s.called = true
	return s.token, s.err
}
