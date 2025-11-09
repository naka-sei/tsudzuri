package page

import "errors"

var (
	ErrNoTitleProvided    = errors.New("no title provided")
	ErrNoUserProvided     = errors.New("no user provided")
	ErrInvalidLinksLength = errors.New("invalid links length")
	ErrNotCreatedByUser   = errors.New("page not created by the user")
)

type NotFoundLinkError struct {
	URL string
}

func (e *NotFoundLinkError) Error() string {
	return "link not found: " + e.URL
}

func ErrNotFoundLink(url string) *NotFoundLinkError {
	return &NotFoundLinkError{URL: url}
}
