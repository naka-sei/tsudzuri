package auth

import (
	"context"
	"errors"
	"testing"

	"firebase.google.com/go/v4/auth"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	duser "github.com/naka-sei/tsudzuri/domain/user"
	mockuser "github.com/naka-sei/tsudzuri/domain/user/mock/mock_user"
	mockauthenticator "github.com/naka-sei/tsudzuri/infrastructure/api/firebase/mock/mock_authenticator"
	cacheiface "github.com/naka-sei/tsudzuri/pkg/cache"
	mockcache "github.com/naka-sei/tsudzuri/pkg/cache/mock/mock_cache"
	ctxuser "github.com/naka-sei/tsudzuri/pkg/ctx/user"
)

func TestNewAuthenticationUnaryServerInterceptor(t *testing.T) {
	email := "test@example.com"
	testUser := duser.ReconstructUser("test-id", "test-uid", "google", &email)
	testToken := &auth.Token{UID: "test-uid"}

	type fields struct {
		authenticator *mockauthenticator.MockAuthenticator
		userRepo      *mockuser.MockUserRepository
		cache         *mockcache.MockCache[*duser.User]
	}

	type args struct {
		ctx      context.Context
		req      interface{}
		info     *grpc.UnaryServerInfo
		useCache bool
	}

	type want struct {
		hasErr          bool
		errCode         codes.Code
		expectUserInCtx bool
	}

	tests := []struct {
		name  string
		setup func(*fields)
		args  args
		want  want
	}{
		{
			name:  "public_method_should_allow_access_without_authentication",
			setup: nil,
			args: args{
				ctx:      context.Background(),
				req:      struct{}{},
				info:     &grpc.UnaryServerInfo{FullMethod: "/tsudzuri.v1.TsudzuriService/CreateUser"},
				useCache: true,
			},
			want: want{
				hasErr:          false,
				expectUserInCtx: false,
			},
		},
		{
			name:  "missing_metadata_should_return_user_not_found_error",
			setup: nil,
			args: args{
				ctx:      context.Background(),
				req:      struct{}{},
				info:     &grpc.UnaryServerInfo{FullMethod: "/tsudzuri.v1.TsudzuriService/GetPage"},
				useCache: true,
			},
			want: want{
				hasErr:  true,
				errCode: codes.Unauthenticated,
			},
		},
		{
			name:  "missing_authorization_header_should_return_user_not_found_error",
			setup: nil,
			args: args{
				ctx:      metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{})),
				req:      struct{}{},
				info:     &grpc.UnaryServerInfo{FullMethod: "/tsudzuri.v1.TsudzuriService/GetPage"},
				useCache: true,
			},
			want: want{
				hasErr:  true,
				errCode: codes.Unauthenticated,
			},
		},
		{
			name: "invalid_token_should_return_user_not_found_error",
			setup: func(f *fields) {
				f.authenticator.EXPECT().
					VerifyIDToken(gomock.Any(), "invalid-token").
					Return(nil, errors.New("invalid token"))
			},
			args: args{
				ctx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
					"authorization": "Bearer invalid-token",
				})),
				req:      struct{}{},
				info:     &grpc.UnaryServerInfo{FullMethod: "/tsudzuri.v1.TsudzuriService/GetPage"},
				useCache: true,
			},
			want: want{
				hasErr:  true,
				errCode: codes.Unauthenticated,
			},
		},
		{
			name: "user_not_found_in_repository_should_return_error",
			setup: func(f *fields) {
				f.authenticator.EXPECT().
					VerifyIDToken(gomock.Any(), "valid-token").
					Return(testToken, nil)
				f.cache.EXPECT().
					Get(gomock.Any(), "test-uid").
					Return(nil, false)
				f.userRepo.EXPECT().
					Get(gomock.Any(), "test-uid").
					Return(nil, duser.ErrUserNotFound)
			},
			args: args{
				ctx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
					"authorization": "Bearer valid-token",
				})),
				req:      struct{}{},
				info:     &grpc.UnaryServerInfo{FullMethod: "/tsudzuri.v1.TsudzuriService/GetPage"},
				useCache: true,
			},
			want: want{
				hasErr:  true,
				errCode: codes.Unauthenticated,
			},
		},
		{
			name: "user_found_but_nil_should_return_user_not_found_error",
			setup: func(f *fields) {
				f.authenticator.EXPECT().
					VerifyIDToken(gomock.Any(), "valid-token").
					Return(testToken, nil)
				f.cache.EXPECT().
					Get(gomock.Any(), "test-uid").
					Return(nil, false)
				f.userRepo.EXPECT().
					Get(gomock.Any(), "test-uid").
					Return(nil, nil)
			},
			args: args{
				ctx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
					"authorization": "Bearer valid-token",
				})),
				req:      struct{}{},
				info:     &grpc.UnaryServerInfo{FullMethod: "/tsudzuri.v1.TsudzuriService/GetPage"},
				useCache: true,
			},
			want: want{
				hasErr:  true,
				errCode: codes.Unauthenticated,
			},
		},
		{
			name: "successful_authentication_with_cache_miss_should_set_cache_and_proceed",
			setup: func(f *fields) {
				f.authenticator.EXPECT().
					VerifyIDToken(gomock.Any(), "valid-token").
					Return(testToken, nil)
				f.cache.EXPECT().
					Get(gomock.Any(), "test-uid").
					Return(nil, false)
				f.userRepo.EXPECT().
					Get(gomock.Any(), "test-uid").
					Return(testUser, nil)
				f.cache.EXPECT().
					Set(gomock.Any(), "test-uid", testUser)
			},
			args: args{
				ctx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
					"authorization": "Bearer valid-token",
				})),
				req:      struct{}{},
				info:     &grpc.UnaryServerInfo{FullMethod: "/tsudzuri.v1.TsudzuriService/GetPage"},
				useCache: true,
			},
			want: want{
				hasErr:          false,
				expectUserInCtx: true,
			},
		},
		{
			name: "successful_authentication_with_cache_hit_should_skip_repository_call",
			setup: func(f *fields) {
				f.authenticator.EXPECT().
					VerifyIDToken(gomock.Any(), "valid-token").
					Return(testToken, nil)
				f.cache.EXPECT().
					Get(gomock.Any(), "test-uid").
					Return(testUser, true)
			},
			args: args{
				ctx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
					"authorization": "Bearer valid-token",
				})),
				req:      struct{}{},
				info:     &grpc.UnaryServerInfo{FullMethod: "/tsudzuri.v1.TsudzuriService/GetPage"},
				useCache: true,
			},
			want: want{
				hasErr:          false,
				expectUserInCtx: true,
			},
		},
		{
			name: "successful_authentication_without_cache_should_work",
			setup: func(f *fields) {
				f.authenticator.EXPECT().
					VerifyIDToken(gomock.Any(), "valid-token").
					Return(testToken, nil)
				f.userRepo.EXPECT().
					Get(gomock.Any(), "test-uid").
					Return(testUser, nil)
			},
			args: args{
				ctx: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
					"authorization": "Bearer valid-token",
				})),
				req:      struct{}{},
				info:     &grpc.UnaryServerInfo{FullMethod: "/tsudzuri.v1.TsudzuriService/GetPage"},
				useCache: false,
			},
			want: want{
				hasErr:          false,
				expectUserInCtx: true,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			f := &fields{
				authenticator: mockauthenticator.NewMockAuthenticator(ctrl),
				userRepo:      mockuser.NewMockUserRepository(ctrl),
				cache:         mockcache.NewMockCache[*duser.User](ctrl),
			}

			if tt.setup != nil {
				tt.setup(f)
			}

			var cacheUnderTest cacheiface.Cache[*duser.User]
			if tt.args.useCache {
				cacheUnderTest = f.cache
			}

			interceptor := NewAuthenticationUnaryServerInterceptor(
				f.authenticator,
				f.userRepo,
				cacheUnderTest,
			)

			handler := func(ctx context.Context, req interface{}) (interface{}, error) {
				if !tt.want.hasErr && tt.want.expectUserInCtx {
					user, ok := ctxuser.UserFromContext(ctx)
					if !ok || user == nil {
						t.Fatalf("expected user in context for authenticated method")
					}
				}
				return "success", nil
			}

			_, err := interceptor(tt.args.ctx, tt.args.req, tt.args.info, handler)

			if tt.want.hasErr {
				if err == nil {
					t.Fatalf("expected error but got none")
				}

				st, ok := status.FromError(err)
				if !ok {
					t.Fatalf("expected gRPC status error, got: %v", err)
				}

				if st.Code() != tt.want.errCode {
					t.Fatalf("expected error code %v, got %v", tt.want.errCode, st.Code())
				}
			} else if err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}
		})
	}
}

func TestExtractAuthorization(t *testing.T) {
	type args struct {
		md metadata.MD
	}

	type want struct {
		token string
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "should_extract_token_from_authorization_header",
			args: args{
				md: metadata.New(map[string]string{
					"authorization": "Bearer test-token",
				}),
			},
			want: want{
				token: "test-token",
			},
		},
		{
			name: "should_extract_token_from_Authorization_header_capitalized",
			args: args{
				md: metadata.New(map[string]string{
					"Authorization": "Bearer test-token",
				}),
			},
			want: want{
				token: "test-token",
			},
		},
		{
			name: "should_return_token_without_Bearer_prefix",
			args: args{
				md: metadata.New(map[string]string{
					"authorization": "test-token-direct",
				}),
			},
			want: want{
				token: "test-token-direct",
			},
		},
		{
			name: "should_handle_case_insensitive_Bearer_prefix",
			args: args{
				md: metadata.New(map[string]string{
					"authorization": "BEARER test-token",
				}),
			},
			want: want{
				token: "test-token",
			},
		},
		{
			name: "should_return_empty_string_when_no_authorization_header",
			args: args{
				md: metadata.New(map[string]string{}),
			},
			want: want{
				token: "",
			},
		},
		{
			name: "should_handle_whitespace_around_token",
			args: args{
				md: metadata.New(map[string]string{
					"authorization": "  Bearer   test-token  ",
				}),
			},
			want: want{
				token: "test-token",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			result := extractAuthorization(tt.args.md)
			if result != tt.want.token {
				t.Fatalf("expected %q, got %q", tt.want.token, result)
			}
		})
	}
}
