package user

import (
	"context"

	duser "github.com/naka-sei/tsudzuri/domain/user"
	"github.com/naka-sei/tsudzuri/pkg/cache"
	"github.com/naka-sei/tsudzuri/pkg/log"
	"github.com/naka-sei/tsudzuri/pkg/trace"
	uuser "github.com/naka-sei/tsudzuri/usecase/user"
)

type CreateRequest struct {
	UID string `json:"uid"`
}

type UserResponse struct {
	ID       string  `json:"id"`
	UID      string  `json:"uid"`
	Provider string  `json:"provider"`
	Email    *string `json:"email"`
}

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

// SetCache configures the user cache that should be populated on successful user creation.
func (s *CreateService) SetCache(c cache.Cache[*duser.User]) {
	s.cache = c
}

// Create is a transport-agnostic presentation handler.
// It expects a context and a request DTO; returns a response DTO or error.
func (s *CreateService) Create(ctx context.Context, req CreateRequest) (*UserResponse, error) {
	ctx, end := trace.StartSpan(ctx, "presentation/http/user.Create")
	defer end()

	l := log.LoggerFromContext(ctx)
	l.Sugar().Infof("User create request: uid=%s", req.UID)

	u, err := s.usecase.create.Create(ctx, req.UID)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, nil
	}

	res := &UserResponse{
		ID:       u.ID(),
		UID:      u.UID(), // Assuming UID method exists
		Provider: string(u.Provider()),
		Email:    u.Email(),
	}

	if s.cache != nil {
		s.cache.Set(ctx, u.UID(), u)
	}
	l.Sugar().Infof("User created: user_uid=%s", u.UID())
	return res, nil
}
