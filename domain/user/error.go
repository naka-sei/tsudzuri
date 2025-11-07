package user

import "fmt"

var (
	ErrNoSpecifiedEmail = fmt.Errorf("no specified email")
	ErrUserNotFound     = fmt.Errorf("user not found")
)

type InvalidUIDError struct {
	uid string
}

func (e *InvalidUIDError) Error() string {
	return fmt.Sprintf("invalid uid: %s", e.uid)
}

func ErrInvalidUID(uid string) *InvalidUIDError {
	return &InvalidUIDError{uid: uid}
}

type InvalidProviderError struct {
	provider Provider
}

func (e *InvalidProviderError) Error() string {
	return fmt.Sprintf("invalid provider: %s", e.provider)
}

func ErrInvalidProvider(provider Provider) *InvalidProviderError {
	return &InvalidProviderError{provider: provider}
}

type AlreadyLoggedInError struct {
	provider Provider
}

func (e *AlreadyLoggedInError) Error() string {
	return fmt.Sprintf("already logged in with provider: %s", e.provider)
}

func ErrAlreadyLoggedIn(provider Provider) *AlreadyLoggedInError {
	return &AlreadyLoggedInError{provider: provider}
}
