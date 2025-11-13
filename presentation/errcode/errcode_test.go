package errcode

import (
	"errors"
	"testing"

	cmp "github.com/google/go-cmp/cmp"
	cmpopts "github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/grpc/codes"

	dpage "github.com/naka-sei/tsudzuri/domain/page"
	duser "github.com/naka-sei/tsudzuri/domain/user"
)

func TestGetErrorReason(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want *ErrorReason
	}{
		{
			name: "nil error",
			err:  nil,
			want: nil,
		},
		{
			name: "page ErrNoTitleProvided",
			err:  dpage.ErrNoTitleProvided,
			want: &ErrorReason{
				ErrorCode: CodePageInvalidParameter,
				Message:   "タイトルを入力してください。",
			},
		},
		{
			name: "page ErrNoUserProvided",
			err:  dpage.ErrNoUserProvided,
			want: &ErrorReason{
				ErrorCode: CodePageInternalError,
				Message:   "ユーザー情報を取得できませんでした。再度お試しください。",
			},
		},
		{
			name: "page ErrInvalidLinksLength",
			err:  dpage.ErrInvalidLinksLength,
			want: &ErrorReason{
				ErrorCode: CodePageInvalidParameter,
				Message:   "リンクの数が現在のページのリンク数と一致しません。再度お試しください。",
			},
		},
		{
			name: "page ErrNotCreatedByUser",
			err:  dpage.ErrNotCreatedByUser,
			want: &ErrorReason{
				ErrorCode: CodePageAuthorizationFailed,
				Message:   "ページの作成者ではないため、操作を実行できません。",
			},
		},
		{
			name: "page NotFoundLinkError",
			err:  &dpage.NotFoundLinkError{URL: "http://example.com"},
			want: &ErrorReason{
				ErrorCode: CodePageInvalidParameter,
				Message:   "ページに存在しないリンクのため、操作を実行できません。リンク: http://example.com",
			},
		},
		{
			name: "user InvalidProviderError",
			err:  duser.ErrInvalidProvider(duser.Provider("invalid")),
			want: &ErrorReason{
				ErrorCode: CodeUserInvalidParameter,
				Message:   "無効な認証プロバイダーのため、ログインできません。",
			},
		},
		{
			name: "user AlreadyLoggedInError",
			err:  duser.ErrAlreadyLoggedIn(duser.Provider("google")),
			want: &ErrorReason{
				ErrorCode: CodeUserAuthorizationFailed,
				Message:   "既に他の認証プロバイダーでログインしているため、ログインできません。",
			},
		},
		{
			name: "user ErrUserNotFound",
			err:  duser.ErrUserNotFound,
			want: &ErrorReason{
				ErrorCode: CodeUserUnauthorized,
				Message:   "認証が必要です。ログインしてください。",
			},
		},
		{
			name: "user ErrNoSpecifiedEmail",
			err:  duser.ErrNoSpecifiedEmail,
			want: &ErrorReason{
				ErrorCode: CodeUserInternalError,
				Message:   "メールアドレスが取得できませんでした。再度お試しください。",
			},
		},
		{
			name: "unknown error",
			err:  errors.New("unknown error"),
			want: &ErrorReason{
				ErrorCode: CodeUnknownError,
				Message:   "不明なエラーが発生しました。再度お試しください。",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetErrorReason(tt.err)
			if diff := cmp.Diff(tt.want, got, cmpopts.IgnoreUnexported(ErrorCode{})); diff != "" {
				t.Errorf("GetErrorReason() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestGetGRPCCode(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want codes.Code
	}{
		{
			name: "invalid argument",
			err:  dpage.ErrNoTitleProvided,
			want: codes.InvalidArgument,
		},
		{
			name: "permission denied",
			err:  dpage.ErrNotCreatedByUser,
			want: codes.PermissionDenied,
		},
		{
			name: "unauthenticated",
			err:  duser.ErrUserNotFound,
			want: codes.Unauthenticated,
		},
		{
			name: "internal",
			err:  errors.New("unknown"),
			want: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getGRPCCode(tt.err); got != tt.want {
				t.Errorf("GetGRPCCode() = %v, want %v", got, tt.want)
			}
		})
	}
}
