package page

import "fmt"

var (
	ErrPageNotFound = fmt.Errorf("page not found")
	ErrUserNotFound = fmt.Errorf("user not found")
)
