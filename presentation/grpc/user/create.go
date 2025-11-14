package user

import (
	"context"

	tsudzuriv1 "github.com/naka-sei/tsudzuri/api/tsudzuri/v1"
	duser "github.com/naka-sei/tsudzuri/domain/user"
	"github.com/naka-sei/tsudzuri/pkg/cache"
	"github.com/naka-sei/tsudzuri/pkg/log"
	"github.com/naka-sei/tsudzuri/pkg/trace"
	uuser "github.com/naka-sei/tsudzuri/usecase/user"
)

type CreateService struct {
	usecase struct {
		create uuser.CreateUsecase
	}
	cache cache.Cache[*duser.User]
}

func NewCreateService(cu uuser.CreateUsecase) *CreateService {
	return &CreateService{
		usecase: struct{ create uuser.CreateUsecase }{create: cu},
	}
}

func (s *CreateService) SetCache(c cache.Cache[*duser.User]) {
	s.cache = c
}

func (s *CreateService) Create(ctx context.Context, req *tsudzuriv1.CreateUserRequest) (*tsudzuriv1.User, error) {
	ctx, end := trace.StartSpan(ctx, "presentation/grpc/user.Create")
	defer end()

	logger := log.LoggerFromContext(ctx)
	logger.Sugar().Infof("User create request: uid=%s", req.GetUid())

	u, err := s.usecase.create.Create(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, nil
	}

	resp := toProtoUser(u)

	if s.cache != nil {
		s.cache.Set(ctx, u.UID(), u)
	}

	logger.Sugar().Infof("User created: user_uid=%s", u.UID())
	return resp, nil
}
