package page

import (
	"context"

	dpage "github.com/naka-sei/tsudzuri/domain/page"
	duser "github.com/naka-sei/tsudzuri/domain/user"
	ctxuser "github.com/naka-sei/tsudzuri/pkg/ctx/user"
	"github.com/naka-sei/tsudzuri/pkg/log"
	"github.com/naka-sei/tsudzuri/pkg/trace"
)

//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -destination mock/mock_get/get.go -source=./get.go -package=mockgetusecase
type GetUsecase interface {
	// Get returns a page by its ID. The user is obtained from context via pkg/ctx/user.UserFromContext.
	Get(ctx context.Context, pageID string) (*dpage.Page, error)
}

type getUsecase struct {
	repository struct {
		page dpage.PageRepository
	}
}

func NewGetUsecase(pageRepo dpage.PageRepository) GetUsecase {
	u := &getUsecase{
		repository: struct {
			page dpage.PageRepository
		}{
			page: pageRepo,
		},
	}
	return u
}

func (u *getUsecase) Get(ctx context.Context, pageID string) (*dpage.Page, error) {
	ctx, end := trace.StartSpan(ctx, "usecase/page/getUsecase.Get")
	defer end()

	l := log.LoggerFromContext(ctx)
	l.Sugar().Infof("Getting page with id: %s", pageID)

	_, ok := ctxuser.UserFromContext(ctx)
	if !ok {
		return nil, duser.ErrUserNotFound
	}

	page, err := u.repository.page.Get(ctx, pageID)
	if err != nil {
		return nil, err
	}

	if page == nil {
		return nil, ErrPageNotFound
	}

	return page, nil
}
