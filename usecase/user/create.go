package user

import (
	"context"

	duser "github.com/naka-sei/tsudzuri/domain/user"
	"github.com/naka-sei/tsudzuri/pkg/log"
	"github.com/naka-sei/tsudzuri/pkg/trace"
	"github.com/naka-sei/tsudzuri/usecase/service"
)

// ...existing code...

// ...existing code...

//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -destination mock/mock_create/create.go -source=./create.go -package=mockcreateusecase

type CreateUsecase interface {
	// Create creates a new user.
	Create(ctx context.Context, uid string) (*duser.User, error)
}

type createUsecase struct {
	repository struct {
		user duser.UserRepository
	}
	service struct {
		txn service.TransactionService
	}
}

func NewCreateUsecase(
	userRepo duser.UserRepository,
	txnService service.TransactionService,
) CreateUsecase {
	u := &createUsecase{
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
	return u
}

func (u *createUsecase) Create(ctx context.Context, uid string) (*duser.User, error) {
	ctx, end := trace.StartSpan(ctx, "usecase/user/createUsecase.Create")
	defer end()

	l := log.LoggerFromContext(ctx)
	l.Sugar().Infof("Creating user uid: %s", uid)

	newUser := duser.NewUser(uid)
	err := u.service.txn.RunInTransaction(ctx, func(ctx context.Context) error {
		var err error
		newUser, err = u.repository.user.Save(ctx, newUser)
		return err
	})
	if err != nil {
		return nil, err
	}
	return newUser, nil
}
