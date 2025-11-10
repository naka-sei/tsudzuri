package page

import (
	"context"

	dpage "github.com/naka-sei/tsudzuri/domain/page"
	duser "github.com/naka-sei/tsudzuri/domain/user"
	ctxuser "github.com/naka-sei/tsudzuri/pkg/ctx/user"
	"github.com/naka-sei/tsudzuri/pkg/log"
	"github.com/naka-sei/tsudzuri/pkg/trace"
	"github.com/naka-sei/tsudzuri/usecase/service"
)

//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -destination mock/mock_edit/edit.go -source=./edit.go -package=mockeditusecase
type EditUsecase interface {
	// Edit edits a page by its ID. The user is obtained from context via pkg/ctx/user.UserFromContext.
	Edit(ctx context.Context, pageID string, title string, links dpage.Links) error
}

type editUsecase struct {
	repository struct {
		page dpage.PageRepository
	}
	service struct {
		txn service.TransactionService
	}
}

func NewEditUsecase(pageRepo dpage.PageRepository, txn service.TransactionService) EditUsecase {
	u := &editUsecase{
		repository: struct{ page dpage.PageRepository }{page: pageRepo},
		service:    struct{ txn service.TransactionService }{txn: txn},
	}
	return u
}

// Edit edits a page.
func (u *editUsecase) Edit(ctx context.Context, pageID string, title string, links dpage.Links) error {
	ctx, end := trace.StartSpan(ctx, "usecase/page/editUsecase.Edit")
	defer end()

	l := log.LoggerFromContext(ctx)
	l.Sugar().Infof("Editing page id: %s title: %s", pageID, title)

	page, err := u.repository.page.Get(ctx, pageID)
	if err != nil {
		return err
	}

	if page == nil {
		return ErrPageNotFound
	}

	user, ok := ctxuser.UserFromContext(ctx)
	if !ok {
		return duser.ErrUserNotFound
	}

	return u.service.txn.RunInTransaction(ctx, func(ctx context.Context) error {
		if err := page.Edit(user, title, links); err != nil {
			return err
		}
		_, err = u.repository.page.Save(ctx, page)
		return err
	})
}
