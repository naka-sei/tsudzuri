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

//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -destination mock/mock_join/join.go -source=./join.go -package=mockjoinusecase
type JoinUsecase interface {
	// Join adds the authenticated user to the page specified by pageID using the invite code.
	Join(ctx context.Context, pageID string, inviteCode string) error
}

type joinUsecase struct {
	repository struct {
		page dpage.PageRepository
	}
	service struct {
		txn service.TransactionService
	}
}

func NewJoinUsecase(
	pageRepo dpage.PageRepository,
	txnService service.TransactionService,
) JoinUsecase {
	return &joinUsecase{
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
}

func (u *joinUsecase) Join(ctx context.Context, pageID string, inviteCode string) error {
	ctx, end := trace.StartSpan(ctx, "usecase/page/joinUsecase.Join")
	defer end()

	logger := log.LoggerFromContext(ctx)
	logger.Sugar().Infof("Joining page: page_id=%s", pageID)

	user, ok := ctxuser.UserFromContext(ctx)
	if !ok {
		return duser.ErrUserNotFound
	}

	page, err := u.repository.page.Get(ctx, pageID)
	if err != nil {
		return err
	}
	if page == nil {
		return ErrPageNotFound
	}

	return u.service.txn.RunInTransaction(ctx, func(ctx context.Context) error {
		if err := page.Join(user, inviteCode); err != nil {
			return err
		}
		_, err := u.repository.page.Save(ctx, page)
		return err
	})
}
