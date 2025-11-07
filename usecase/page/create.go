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

//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -destination mock/mock_create/create.go -source=./create.go -package=mockcreateusecase
type CreateUsecase interface {
	// Create creates a new page. The user is obtained from context via pkg/ctx/user.UserFromContext.
	Create(ctx context.Context, title string) (*dpage.Page, error)
}

type createUsecase struct {
	repository struct {
		page dpage.PageRepository
	}
	service struct {
		txn service.TransactionService
	}
}

func NewCreateUsecase(
	pageRepo dpage.PageRepository,
	txnService service.TransactionService,
) CreateUsecase {
	u := &createUsecase{
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

// Create creates a new page.
func (u *createUsecase) Create(ctx context.Context, title string) (*dpage.Page, error) {
	ctx, end := trace.StartSpan(ctx, "usecase/page/createUsecase.Create")
	defer end()

	l := log.LoggerFromContext(ctx)
	l.Info(fmt.Sprintf("Creating a new page with title: %s", title))

	user, ok := ctxuser.UserFromContext(ctx)
	if !ok {
		return nil, duser.ErrUserNotFound
	}

	page, err := dpage.NewPage(title, user)
	if err != nil {
		return nil, err
	}

	err = u.service.txn.RunInTransaction(ctx, func(ctx context.Context) error {
		return u.repository.page.Save(ctx, page)
	})
	if err != nil {
		return nil, err
	}

	return page, nil
}
