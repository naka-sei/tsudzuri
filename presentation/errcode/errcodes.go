package errcode

// Error codes for the page domain.
var (
	CodePageInvalidParameter    = newErrorCode("page", "invalid-parameter", "Invalid parameter is provided.")
	CodePageAuthorizationFailed = newErrorCode("page", "authorization-failed", "Authorization failed for the requested operation.")
	CodePageInternalError       = newErrorCode("page", "internal-error", "An internal error occurred in the page domain.")
)

// Error codes for the user domain.
var (
	CodeUserInvalidParameter    = newErrorCode("user", "invalid-parameter", "Invalid parameter is provided.")
	CodeUserAuthorizationFailed = newErrorCode("user", "authorization-failed", "Authorization failed for the requested operation.")
	CodeUserInternalError       = newErrorCode("user", "internal-error", "An internal error occurred in the user domain.")
	CodeUserUnauthorized        = newErrorCode("user", "unauthorized", "Authentication is required or has failed.")
)

var CodeUnknownError = newErrorCode("unknown", "unknown_error", "An unknown error occurred.")
