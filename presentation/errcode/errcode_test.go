package errcode

import (
	"errors"
	"testing"

	cmp "github.com/google/go-cmp/cmp"
	cmpopts "github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/grpc/codes"

	dpage "github.com/naka-sei/tsudzuri/domain/page"
	duser "github.com/naka-sei/tsudzuri/domain/user"
	upage "github.com/naka-sei/tsudzuri/usecase/page"
	uuser "github.com/naka-sei/tsudzuri/usecase/user"
)

func TestGetErrorReason(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want *ErrorReason
	}{
		{
			name: "nil_error",
			err:  nil,
			want: nil,
		},
		{
			name: "page_ErrNoTitleProvided",
			err:  dpage.ErrNoTitleProvided,
			want: &ErrorReason{
				ErrorCode: CodePageInvalidParameter,
				Message:   "タイトルを入力してください。",
			},
		},
		{
			name: "page_ErrNoUserProvided",
			err:  dpage.ErrNoUserProvided,
			want: &ErrorReason{
				ErrorCode: CodePageInternalError,
				Message:   "ユーザー情報を取得できませんでした。再度お試しください。",
			},
		},
		{
			name: "page_ErrInvalidLinksLength",
			err:  dpage.ErrInvalidLinksLength,
			want: &ErrorReason{
				ErrorCode: CodePageInvalidParameter,
				Message:   "リンクの数が現在のページのリンク数と一致しません。再度お試しください。",
			},
		},
		{
			name: "page_ErrNotCreatedByUser",
			err:  dpage.ErrNotCreatedByUser,
			want: &ErrorReason{
				ErrorCode: CodePageAuthorizationFailed,
				Message:   "ページの作成者ではないため、操作を実行できません。",
			},
		},
		{
			name: "page_ErrInvalidInviteCode",
			err:  dpage.ErrInvalidInviteCode,
			want: &ErrorReason{
				ErrorCode: CodePageInvalidParameter,
				Message:   "招待コードが正しくないため、ページに参加できません。",
			},
		},
		{
			name: "page_ErrAlreadyJoined",
			err:  dpage.ErrAlreadyJoined,
			want: &ErrorReason{
				ErrorCode: CodePageInvalidParameter,
				Message:   "すでにこのページに参加しています。",
			},
		},
		{
			name: "page_ErrCreatorCannotJoin",
			err:  dpage.ErrCreatorCannotJoin,
			want: &ErrorReason{
				ErrorCode: CodePageInvalidParameter,
				Message:   "ページ作成者は招待コードによる参加を行う必要はありません。",
			},
		},
		{
			name: "page_NotFoundError",
			err:  upage.ErrPageNotFound,
			want: &ErrorReason{
				ErrorCode: CodePageInvalidParameter,
				Message:   "指定されたページが見つかりません。ページIDを確認してください。",
			},
		},
		{
			name: "page_NotFoundLinkError",
			err:  &dpage.NotFoundLinkError{URL: "http://example.com"},
			want: &ErrorReason{
				ErrorCode: CodePageInvalidParameter,
				Message:   "ページに存在しないリンクのため、操作を実行できません。リンク: http://example.com",
			},
		},
		{
			name: "user_InvalidProviderError",
			err:  duser.ErrInvalidProvider(duser.Provider("invalid")),
			want: &ErrorReason{
				ErrorCode: CodeUserInvalidParameter,
				Message:   "無効な認証プロバイダーのため、ログインできません。",
			},
		},
		{
			name: "user_AlreadyLoggedInError",
			err:  duser.ErrAlreadyLoggedIn(duser.Provider("google")),
			want: &ErrorReason{
				ErrorCode: CodeUserAuthorizationFailed,
				Message:   "既に他の認証プロバイダーでログインしているため、ログインできません。",
			},
		},
		{
			name: "user_ErrUserNotFound",
			err:  duser.ErrUserNotFound,
			want: &ErrorReason{
				ErrorCode: CodeUserUnauthorized,
				Message:   "認証が必要です。ログインしてください。",
			},
		},
		{
			name: "user_ErrNoSpecifiedEmail",
			err:  duser.ErrNoSpecifiedEmail,
			want: &ErrorReason{
				ErrorCode: CodeUserInternalError,
				Message:   "メールアドレスが取得できませんでした。再度お試しください。",
			},
		},
		{
			name: "user_ExistingUserError",
			err:  uuser.ErrExistingUser,
			want: &ErrorReason{
				ErrorCode: CodeUserInvalidParameter,
				Message:   "指定されたユーザーは既に存在しています。",
			},
		},
		{
			name: "unknown_error",
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
			name: "invalid_argument",
			err:  dpage.ErrNoTitleProvided,
			want: codes.InvalidArgument,
		},
		{
			name: "permission_denied",
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
