package user

import (
	"context"

	duser "github.com/naka-sei/tsudzuri/domain/user"
	ctxuser "github.com/naka-sei/tsudzuri/pkg/ctx/user"
	"github.com/naka-sei/tsudzuri/pkg/log"
	"github.com/naka-sei/tsudzuri/pkg/trace"
	"github.com/naka-sei/tsudzuri/usecase/service"
)

//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -destination mock/mock_login/login.go -source=./login.go -package=mockloginusecase

type LoginUsecase interface {
	// Login authenticates and updates the user with provider and email.
	Login(ctx context.Context, provider string, email *string) error
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

func (u *loginUsecase) Login(ctx context.Context, provider string, email *string) error {
	ctx, end := trace.StartSpan(ctx, "usecase/user/loginUsecase.Login")
	defer end()

	l := log.LoggerFromContext(ctx)

	user, ok := ctxuser.UserFromContext(ctx)
	if !ok {
		return duser.ErrUserNotFound
	}

	l.Sugar().Infof("Logging in user id: %s with provider: %s email: %v", user.ID(), provider, email)

	if err := user.Login(provider, email); err != nil {
		return err
	}
	if err := u.service.txn.RunInTransaction(ctx, func(ctx context.Context) error {
		var err error
		user, err = u.repository.user.Save(ctx, user)
		return err
	}); err != nil {
		return err
	}
	return nil
}
