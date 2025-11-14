package user

import (
	"context"

	tsudzuriv1 "github.com/naka-sei/tsudzuri/api/tsudzuri/v1"
	duser "github.com/naka-sei/tsudzuri/domain/user"
	ctxuser "github.com/naka-sei/tsudzuri/pkg/ctx/user"
	"github.com/naka-sei/tsudzuri/pkg/log"
	"github.com/naka-sei/tsudzuri/pkg/trace"
	uuser "github.com/naka-sei/tsudzuri/usecase/user"
	"google.golang.org/protobuf/types/known/emptypb"
)

type GetService struct {
	usecase struct {
		get uuser.GetUsecase
	}
}

func NewGetService(gu uuser.GetUsecase) *GetService {
	return &GetService{
		usecase: struct{ get uuser.GetUsecase }{get: gu},
	}
}

func (s *GetService) Get(ctx context.Context, _ *emptypb.Empty) (*tsudzuriv1.User, error) {
	ctx, end := trace.StartSpan(ctx, "presentation/grpc/user.Get")
	defer end()

	logger := log.LoggerFromContext(ctx)

	userInCtx, ok := ctxuser.UserFromContext(ctx)
	if !ok {
		return nil, duser.ErrUserNotFound
	}

	logger.Sugar().Infof("User get request: uid=%s", userInCtx.UID())

	user, err := s.usecase.get.Get(ctx)
	if err != nil {
		return nil, err
	}

	logger.Sugar().Infof("User get response: uid=%s joined_pages=%d", user.UID(), len(user.JoinedPageIDs()))

	return toProtoUser(user), nil
}
