package user

import (
	"context"

	"github.com/naka-sei/tsudzuri/pkg/log"
	"github.com/naka-sei/tsudzuri/pkg/trace"
	"github.com/naka-sei/tsudzuri/presentation/http/response"
	uuser "github.com/naka-sei/tsudzuri/usecase/user"
)

type LoginRequest struct {
	Provider string  `json:"provider"`
	Email    *string `json:"email"`
}

type LoginService struct {
	usecase struct {
		login uuser.LoginUsecase
	}
}

func NewLoginService(lu uuser.LoginUsecase) *LoginService {
	return &LoginService{
		usecase: struct{ login uuser.LoginUsecase }{login: lu},
	}
}

// Login is a transport-agnostic presentation handler.
// It expects a context and a request DTO; returns a response DTO or error.
func (s *LoginService) Login(ctx context.Context, req LoginRequest) (response.EmptyResponse, error) {
	ctx, end := trace.StartSpan(ctx, "presentation/http/user.Login")
	defer end()

	l := log.LoggerFromContext(ctx)
	l.Sugar().Infof("User login request: provider=%s email=%v", req.Provider, req.Email)

	err := s.usecase.login.Login(ctx, req.Provider, req.Email)
	if err != nil {
		return response.EmptyResponse{}, err
	}

	l.Sugar().Infof("User logged in: provider=%s", req.Provider)
	return response.EmptyResponse{}, nil
}
