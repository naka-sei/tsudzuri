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

type LinkRemoveUsecaseInput struct {
	PageID string
	URL    string
}

//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -destination mock/mock_link_remove/link_remove.go -source=./link_remove.go -package=mocklinkremoveusecase
type LinkRemoveUseCase interface {
	// LinkRemove removes a link from a page. The user is obtained from context via pkg/ctx/user.UserFromContext.
	LinkRemove(ctx context.Context, input LinkRemoveUsecaseInput) error
}

type linkRemoveUsecase struct {
	repository struct {
		page dpage.PageRepository
	}
	service struct {
		txn service.TransactionService
	}
}

func NewLinkRemoveUsecase(
	pageRepo dpage.PageRepository,
	txnService service.TransactionService,
) LinkRemoveUseCase {
	u := &linkRemoveUsecase{
		repository: struct {
			page dpage.PageRepository
		}{
			page: pageRepo,
		},
		service: struct {
			txn service.TransactionService
		}{
			txn: txnService,
		},
	}
	return u
}

func (u *linkRemoveUsecase) LinkRemove(ctx context.Context, input LinkRemoveUsecaseInput) error {
	ctx, end := trace.StartSpan(ctx, "usecase/page/linkRemoveUsecase.LinkRemove")
	defer end()

	l := log.LoggerFromContext(ctx)
	l.Sugar().Infof("Removing link from page %s: url=%s", input.PageID, input.URL)

	page, err := u.repository.page.Get(ctx, input.PageID)
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

	if err := page.Authorize(user); err != nil {
		return err
	}

	return u.service.txn.RunInTransaction(ctx, func(ctx context.Context) error {
		if err := page.RemoveLink(user, input.URL); err != nil {
			return err
		}
		return u.repository.page.Save(ctx, page)
	})
}
