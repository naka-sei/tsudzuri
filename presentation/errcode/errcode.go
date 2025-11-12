package errcode

import (
	"errors"
	"fmt"
	"net/http"

	dpage "github.com/naka-sei/tsudzuri/domain/page"
	duser "github.com/naka-sei/tsudzuri/domain/user"
	"google.golang.org/grpc/codes"
)

type ErrorCode struct {
	domain      string
	code        string
	description string
}

type ErrorReason struct {
	ErrorCode *ErrorCode
	Message   string
}

func newErrorCode(domain, code, description string) *ErrorCode {
	return &ErrorCode{
		domain:      domain,
		code:        code,
		description: description,
	}
}

func GetErrorReason(err error) *ErrorReason {
	if err == nil {
		return nil
	}

	var pageUserErr *dpage.NotFoundLinkError

	switch {
	case errors.Is(err, dpage.ErrNoTitleProvided):
		return &ErrorReason{
			ErrorCode: CodePageInvalidParameter,
			Message:   "タイトルを入力してください。",
		}
	case errors.Is(err, dpage.ErrNoUserProvided):
		return &ErrorReason{
			ErrorCode: CodePageInternalError,
			Message:   "ユーザー情報を取得できませんでした。再度お試しください。",
		}
	case errors.Is(err, dpage.ErrInvalidLinksLength):
		return &ErrorReason{
			ErrorCode: CodePageInvalidParameter,
			Message:   "リンクの数が現在のページのリンク数と一致しません。再度お試しください。",
		}
	case errors.Is(err, dpage.ErrNotCreatedByUser):
		return &ErrorReason{
			ErrorCode: CodePageAuthorizationFailed,
			Message:   "ページの作成者ではないため、操作を実行できません。",
		}
	case errors.As(err, &pageUserErr):
		return &ErrorReason{
			ErrorCode: CodePageInvalidParameter,
			Message:   fmt.Sprintf("ページに存在しないリンクのため、操作を実行できません。リンク: %s", pageUserErr.URL),
		}
	}

	var (
		userProviderErr *duser.InvalidProviderError
		userLoginErr    *duser.AlreadyLoggedInError
	)

	switch {
	case errors.As(err, &userProviderErr):
		return &ErrorReason{
			ErrorCode: CodeUserInvalidParameter,
			Message:   "無効な認証プロバイダーのため、ログインできません。",
		}
	case errors.As(err, &userLoginErr):
		return &ErrorReason{
			ErrorCode: CodeUserAuthorizationFailed,
			Message:   "既に他の認証プロバイダーでログインしているため、ログインできません。",
		}
	case errors.Is(err, duser.ErrUserNotFound):
		return &ErrorReason{
			ErrorCode: CodeUserUnauthorized,
			Message:   "認証が必要です。ログインしてください。",
		}
	case errors.Is(err, duser.ErrNoSpecifiedEmail):
		return &ErrorReason{
			ErrorCode: CodeUserInternalError,
			Message:   "メールアドレスが取得できませんでした。再度お試しください。",
		}
	}

	return &ErrorReason{
		ErrorCode: CodeUnknownError,
		Message:   "不明なエラーが発生しました。再度お試しください。",
	}
}

// GetStatusCode maps error codes to HTTP status codes.
func GetStatusCode(err error) int {
	reason := GetErrorReason(err)
	switch reason.ErrorCode {
	case CodePageInvalidParameter, CodeUserInvalidParameter:
		return http.StatusBadRequest
	case CodePageAuthorizationFailed, CodeUserAuthorizationFailed:
		return http.StatusForbidden
	case CodePageInternalError, CodeUserInternalError:
		return http.StatusInternalServerError
	case CodeUserUnauthorized:
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}

// GetGRPCCode maps error codes to gRPC status codes.
func GetGRPCCode(err error) codes.Code {
	switch GetStatusCode(err) {
	case http.StatusBadRequest:
		return codes.InvalidArgument
	case http.StatusForbidden:
		return codes.PermissionDenied
	case http.StatusUnauthorized:
		return codes.Unauthenticated
	case http.StatusInternalServerError:
		return codes.Internal
	default:
		return codes.Unknown
	}
}
