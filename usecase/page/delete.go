package page

import (
	"context"

	dpage "github.com/naka-sei/tsudzuri/domain/page"
	ctxuser "github.com/naka-sei/tsudzuri/pkg/ctx/user"
	"github.com/naka-sei/tsudzuri/usecase/service"
)

//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -destination mock/mock_delete/delete.go -source=./delete.go -package=mockdeleteusecase
type DeleteUsecase interface {
	// Delete deletes a page by its ID. The user is obtained from context via pkg/ctx/user.UserFromContext.
	Delete(ctx context.Context, pageID string) error
}

type deleteUsecase struct {
	repository struct {
		page dpage.PageRepository
	}
	service struct {
		txn service.TransactionService
	}
}

func NewDeleteUsecase(
	pageRepo dpage.PageRepository,
	txnService service.TransactionService,
) DeleteUsecase {
	u := &deleteUsecase{
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

// Delete deletes a page by its ID.
func (u *deleteUsecase) Delete(ctx context.Context, pageID string) error {
	page, err := u.repository.page.Get(ctx, pageID)
	if err != nil {
		return err
	}

	if page == nil {
		return ErrPageNotFound
	}

	user, ok := ctxuser.UserFromContext(ctx)
	if !ok {
		return ErrUserNotFound
	}

	if err := page.Authorize(user); err != nil {
		return err
	}

	return u.service.txn.RunInTransaction(ctx, func(ctx context.Context) error {
		return u.repository.page.DeleteByID(ctx, pageID)
	})
}
