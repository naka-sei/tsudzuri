package user

import (
	"context"

	duser "github.com/naka-sei/tsudzuri/domain/user"
	ctxuser "github.com/naka-sei/tsudzuri/pkg/ctx/user"
	"github.com/naka-sei/tsudzuri/pkg/log"
	"github.com/naka-sei/tsudzuri/pkg/trace"
)

//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -destination mock/mock_get/get.go -source=./get.go -package=mockgetusecase

type GetUsecase interface {
	// Get returns the authenticated user details.
	Get(ctx context.Context) (*duser.User, error)
}

type getUsecase struct {
	repository struct {
		user duser.UserRepository
	}
}

func NewGetUsecase(userRepo duser.UserRepository) GetUsecase {
	return &getUsecase{
		repository: struct{ user duser.UserRepository }{user: userRepo},
	}
}

func (u *getUsecase) Get(ctx context.Context) (*duser.User, error) {
	ctx, end := trace.StartSpan(ctx, "usecase/user/getUsecase.Get")
	defer end()

	logger := log.LoggerFromContext(ctx)

	currentUser, ok := ctxuser.UserFromContext(ctx)
	if !ok {
		return nil, duser.ErrUserNotFound
	}

	logger.Sugar().Infof("Fetching user details uid=%s", currentUser.UID())

	fetched, err := u.repository.user.Get(ctx, currentUser.UID())
	if err != nil {
		return nil, err
	}
	if fetched == nil {
		return nil, duser.ErrUserNotFound
	}

	return fetched, nil
}
