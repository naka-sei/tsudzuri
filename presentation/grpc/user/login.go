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

func (s *LoginService) Login(ctx context.Context, req *tsudzuriv1.LoginRequest) (*emptypb.Empty, error) {
	ctx, end := trace.StartSpan(ctx, "presentation/grpc/user.Login")
	defer end()

	logger := log.LoggerFromContext(ctx)
	user, ok := ctxuser.UserFromContext(ctx)
	if !ok {
		return nil, duser.ErrUserNotFound
	}

	emailValue := req.GetEmail()
	logger.Sugar().Infof("User login request: provider=%s email=%v user_uid=%s", req.GetProvider(), emailValue, user.UID())

	var email *string
	if emailValue != nil {
		email = &emailValue.Value
	}

	if err := s.usecase.login.Login(ctx, req.GetProvider(), email); err != nil {
		return nil, err
	}

	logger.Sugar().Infof("User logged in: provider=%s user_uid=%s", req.GetProvider(), user.UID())
	return &emptypb.Empty{}, nil
}
