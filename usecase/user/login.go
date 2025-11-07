package user

import (
	"context"

	duser "github.com/naka-sei/tsudzuri/domain/user"
	ctxuser "github.com/naka-sei/tsudzuri/pkg/ctx/user"
	"github.com/naka-sei/tsudzuri/usecase/service"
)

//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -destination mock/mock_login/login.go -source=./login.go -package=mockloginusecase

type LoginUsecase interface {
	// Login authenticates and updates the user with provider and email.
	Login(ctx context.Context, uid string, provider string, email *string) error
}

type loginUsecase struct {
	repository struct {
		user duser.UserRepository
	}
	service struct {
		txn service.TransactionService
	}
}

func NewLoginUsecase(userRepo duser.UserRepository, txnService service.TransactionService) LoginUsecase {
	return &loginUsecase{
		repository: struct {
			user duser.UserRepository
		}{
			user: userRepo,
		},
		service: struct {
			txn service.TransactionService
		}{
			txn: txnService,
		},
	}
}

func (u *loginUsecase) Login(ctx context.Context, uid string, provider string, email *string) error {
	user, ok := ctxuser.UserFromContext(ctx)
	if !ok {
		return duser.ErrUserNotFound
	}
	if err := user.Login(uid, provider, email); err != nil {
		return err
	}
	if err := u.service.txn.RunInTransaction(ctx, func(ctx context.Context) error {
		return u.repository.user.Save(ctx, user)
	}); err != nil {
		return err
	}
	return nil
}
