package page

import "errors"

var (
	ErrInvalidLinksLength = errors.New("invalid links length")
	ErrNotCreatedByUser   = errors.New("page not created by the user")
	ErrNoUserProvided     = errors.New("no user provided")
)

type NotFoundLinkError struct {
	url string
}

func (e *NotFoundLinkError) Error() string {
	return "link not found: " + e.url
}

func ErrNotFoundLink(url string) *NotFoundLinkError {
	return &NotFoundLinkError{url: url}
}
