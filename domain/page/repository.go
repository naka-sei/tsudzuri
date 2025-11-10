package page

import "context"

//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -destination mock/mock_page/page.go -source=./repository.go -package=mockpage

type PageRepository interface {
	Get(ctx context.Context, id string) (*Page, error)
	List(ctx context.Context, options ...SearchOption) ([]*Page, error)
	Save(ctx context.Context, page *Page) (*Page, error)
	DeleteByID(ctx context.Context, id string) error
}

type SearchParams struct {
	IDs             []string
	CreatedByUserID string
	Page            *int32
	PageSize        *int32
}

type SearchOption interface {
	Apply(*SearchParams)
}

type optionFunc func(*SearchParams)

func (f optionFunc) Apply(p *SearchParams) {
	f(p)
}

func WithIDs(ids []string) SearchOption {
	return optionFunc(func(p *SearchParams) {
		p.IDs = ids
	})
}

func WithCreatedByUserID(userID string) SearchOption {
	return optionFunc(func(p *SearchParams) {
		p.CreatedByUserID = userID
	})
}

func WithPageSearchOption(page int32) SearchOption {
	return optionFunc(func(p *SearchParams) {
		p.Page = &page
	})
}

func WithPageSizeSearchOption(pageSize int32) SearchOption {
	return optionFunc(func(p *SearchParams) {
		p.PageSize = &pageSize
	})
}
