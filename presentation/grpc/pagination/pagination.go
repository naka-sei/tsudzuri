package pagination

import "fmt"

var ErrInvalidPage = fmt.Errorf("invalid page")

const (
	defaultPage     = 1
	minPage         = 1
	DefaultPageSize = 20
	MaxPageSize     = 100
	minPageSize     = 1
)

// ValidtePage validates the page number.
func ValidtePage(page *int32) (int32, error) {
	if page == nil {
		return defaultPage, nil
	}
	if *page < minPage {
		return 0, fmt.Errorf("page must be at least %d: %w", minPage, ErrInvalidPage)
	}
	return *page, nil
}

// ValidatePageSize validates the page size.
func ValidatePageSize(pageSize *int32) (int32, error) {
	if pageSize == nil {
		return DefaultPageSize, nil
	}
	if *pageSize < minPageSize {
		return 0, fmt.Errorf("page size must be at least %d: %w", minPageSize, ErrInvalidPage)
	}
	if *pageSize > MaxPageSize {
		return 0, fmt.Errorf("page size must be at most %d: %w", MaxPageSize, ErrInvalidPage)
	}
	return *pageSize, nil
}
