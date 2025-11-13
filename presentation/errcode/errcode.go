package errcode

import (
	"errors"
	"fmt"

	dpage "github.com/naka-sei/tsudzuri/domain/page"
	duser "github.com/naka-sei/tsudzuri/domain/user"
	upage "github.com/naka-sei/tsudzuri/usecase/page"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	case errors.Is(err, upage.ErrPageNotFound):
		return &ErrorReason{
			ErrorCode: CodePageInvalidParameter,
			Message:   "指定されたページが見つかりません。ページIDを確認してください。",
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

// getGRPCCode maps error codes to gRPC status codes.
func getGRPCCode(err error) codes.Code {
	errReason := GetErrorReason(err)
	switch errReason.ErrorCode {
	case CodeUserUnauthorized:
		return codes.Unauthenticated
	case CodeUserAuthorizationFailed, CodePageAuthorizationFailed:
		return codes.PermissionDenied
	case CodeUserInvalidParameter, CodePageInvalidParameter:
		return codes.InvalidArgument
	case CodeUserInternalError, CodePageInternalError:
		return codes.Internal
	case CodeUnknownError:
		return codes.Internal
	default:
		return codes.Unknown
	}
}

// ToGRPCStatus converts the given error into a gRPC status error with a user-facing message.
// If the error is already a gRPC status, it is returned as-is to preserve additional details.
func ToGRPCStatus(err error) error {
	if err == nil {
		return nil
	}

	if _, ok := status.FromError(err); ok {
		return err
	}

	code := getGRPCCode(err)
	reason := GetErrorReason(err)

	st, err := status.New(code, reason.ErrorCode.description).WithDetails(&errdetails.ErrorInfo{
		Reason:   reason.ErrorCode.code,
		Metadata: map[string]string{"client_message": reason.Message},
	})
	if err != nil {
		return status.New(code, reason.Message).Err()
	}

	return st.Err()
}
