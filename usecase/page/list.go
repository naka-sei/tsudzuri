package page

import (
	"context"

	dpage "github.com/naka-sei/tsudzuri/domain/page"
	duser "github.com/naka-sei/tsudzuri/domain/user"
	ctxuser "github.com/naka-sei/tsudzuri/pkg/ctx/user"
)

//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -destination mock/mock_list/list.go -source=./list.go -package=mocklistusecase
type ListUsecase interface {
	// List returns a list of pages. The user is obtained from context via pkg/ctx/user.UserFromContext.
	List(ctx context.Context, options ...dpage.SearchOption) ([]*dpage.Page, error)
}

type listUsecase struct {
	repository struct {
		page dpage.PageRepository
	}
}

func NewListUsecase(pageRepo dpage.PageRepository) ListUsecase {
	u := &listUsecase{
		repository: struct {
			page dpage.PageRepository
		}{
			page: pageRepo,
		},
	}
	return u
}

func (u *listUsecase) List(ctx context.Context, options ...dpage.SearchOption) ([]*dpage.Page, error) {
	user, ok := ctxuser.UserFromContext(ctx)
	if !ok {
		return nil, duser.ErrUserNotFound
	}

	pages, err := u.repository.page.List(ctx, user.ID(), options...)
	if err != nil {
		return nil, err
	}

	return pages, nil
}
