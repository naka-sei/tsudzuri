package page

import (
	"context"
	"fmt"

	dpage "github.com/naka-sei/tsudzuri/domain/page"
	duser "github.com/naka-sei/tsudzuri/domain/user"
	ctxuser "github.com/naka-sei/tsudzuri/pkg/ctx/user"
	"github.com/naka-sei/tsudzuri/pkg/log"
	"github.com/naka-sei/tsudzuri/pkg/trace"
	"github.com/naka-sei/tsudzuri/usecase/service"
)

type LinkAddUsecaseInput struct {
	PageID string
	URL    string
	Memo   string
}

//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -destination mock/mock_link_add/link_add.go -source=./link_add.go -package=mocklinkaddusecase
type LinkAddUseCase interface {
	// LinkAdd adds a link to a page. The user is obtained from context via pkg/ctx/user.UserFromContext.
	LinkAdd(ctx context.Context, input LinkAddUsecaseInput) error
}

type linkAddUsecase struct {
	repository struct {
		page dpage.PageRepository
	}
	service struct {
		txn service.TransactionService
	}
}

func NewLinkAddUsecase(
	pageRepo dpage.PageRepository,
	txnService service.TransactionService,
) LinkAddUseCase {
	u := &linkAddUsecase{
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

func (u *linkAddUsecase) LinkAdd(ctx context.Context, input LinkAddUsecaseInput) error {
	ctx, end := trace.StartSpan(ctx, "usecase/page/linkAddUsecase.LinkAdd")
	defer end()

	l := log.LoggerFromContext(ctx)
	l.Info(fmt.Sprintf("Adding link to page %s: url=%s memo=%s", input.PageID, input.URL, input.Memo))

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
		if err := page.AddLink(user, input.URL, input.Memo); err != nil {
			return err
		}
		return u.repository.page.Save(ctx, page)
	})
}
