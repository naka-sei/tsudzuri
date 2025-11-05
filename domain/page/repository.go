package page

//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -destination mock/mock_page/page.go -source=./repository.go -package=mockpage

type PageRepository interface {
	FindByID(id int64) (*Page, error)
	Save(page *Page) error
}
