package user

import (
	"context"

	duser "github.com/naka-sei/tsudzuri/domain/user"
	ctxuser "github.com/naka-sei/tsudzuri/pkg/ctx/user"
	"github.com/naka-sei/tsudzuri/pkg/log"
	"github.com/naka-sei/tsudzuri/pkg/trace"
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
func (s *LoginService) Login(ctx context.Context, req LoginRequest) error {
	ctx, end := trace.StartSpan(ctx, "presentation/http/user.Login")
	defer end()

	l := log.LoggerFromContext(ctx)
	u, ok := ctxuser.UserFromContext(ctx)
	if !ok {
		return duser.ErrUserNotFound
	}
	uid := u.UID()
	l.Sugar().Infof("User login request: provider=%s email=%v user_uid=%s", req.Provider, req.Email, uid)

	err := s.usecase.login.Login(ctx, req.Provider, req.Email)
	if err != nil {
		return err
	}

	l.Sugar().Infof("User logged in: provider=%s user_uid=%s", req.Provider, uid)
	return nil
}
