package user

import (
	"context"

	tsudzuriv1 "github.com/naka-sei/tsudzuri/api/tsudzuri/v1"
	duser "github.com/naka-sei/tsudzuri/domain/user"
	"github.com/naka-sei/tsudzuri/pkg/cache"
	"github.com/naka-sei/tsudzuri/presentation/errcode"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Server struct {
	tsudzuriv1.UnimplementedUserServiceServer

	create *CreateService
	login  *LoginService
}

func NewServer(create *CreateService, login *LoginService) *Server {
	return &Server{
		create: create,
		login:  login,
	}
}

func (s *Server) WithUserCache(c cache.Cache[*duser.User]) {
	if s.create != nil {
		s.create.SetCache(c)
	}
}

func (s *Server) CreateUser(ctx context.Context, req *tsudzuriv1.CreateUserRequest) (*tsudzuriv1.User, error) {
	return errcode.WrapGRPC(s.create.Create(ctx, req))
}

func (s *Server) Login(ctx context.Context, req *tsudzuriv1.LoginRequest) (*emptypb.Empty, error) {
	return errcode.WrapGRPC(s.login.Login(ctx, req))
}
