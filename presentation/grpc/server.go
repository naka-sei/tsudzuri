package presentationgrpc

import (
	"google.golang.org/grpc"

	tsudzuriv1 "github.com/naka-sei/tsudzuri/api/tsudzuri/v1"
	duser "github.com/naka-sei/tsudzuri/domain/user"
	"github.com/naka-sei/tsudzuri/pkg/cache"
	grpcpage "github.com/naka-sei/tsudzuri/presentation/grpc/page"
	grpcuser "github.com/naka-sei/tsudzuri/presentation/grpc/user"
)

type Server struct {
	page *grpcpage.Server
	user *grpcuser.Server
}

func NewServer(page *grpcpage.Server, user *grpcuser.Server) *Server {
	return &Server{
		page: page,
		user: user,
	}
}

func (s *Server) RegisterGRPC(registrar grpc.ServiceRegistrar) {
	tsudzuriv1.RegisterPageServiceServer(registrar, s.page)
	tsudzuriv1.RegisterUserServiceServer(registrar, s.user)
}

func (s *Server) WithUserCache(c cache.Cache[*duser.User]) {
	if s == nil {
		return
	}
	if s.user != nil {
		s.user.WithUserCache(c)
	}
}

func (s *Server) PageServer() *grpcpage.Server {
	return s.page
}

func (s *Server) UserServer() *grpcuser.Server {
	return s.user
}
