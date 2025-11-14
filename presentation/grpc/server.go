package presentationgrpc

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"

	tsudzuriv1 "github.com/naka-sei/tsudzuri/api/tsudzuri/v1"
	duser "github.com/naka-sei/tsudzuri/domain/user"
	"github.com/naka-sei/tsudzuri/pkg/cache"
	"github.com/naka-sei/tsudzuri/presentation/errcode"
	grpcpage "github.com/naka-sei/tsudzuri/presentation/grpc/page"
	grpcuser "github.com/naka-sei/tsudzuri/presentation/grpc/user"
)

type Server struct {
	tsudzuriv1.UnimplementedTsudzuriServiceServer

	page struct {
		create     *grpcpage.CreateService
		get        *grpcpage.GetService
		list       *grpcpage.ListService
		edit       *grpcpage.EditService
		delete     *grpcpage.DeleteService
		linkAdd    *grpcpage.LinkAddService
		linkRemove *grpcpage.LinkRemoveService
		join       *grpcpage.JoinService
	}

	user struct {
		create *grpcuser.CreateService
		login  *grpcuser.LoginService
		get    *grpcuser.GetService
	}
}

func NewServer(
	createPage *grpcpage.CreateService,
	getPage *grpcpage.GetService,
	listPages *grpcpage.ListService,
	editPage *grpcpage.EditService,
	deletePage *grpcpage.DeleteService,
	addLink *grpcpage.LinkAddService,
	removeLink *grpcpage.LinkRemoveService,
	joinPage *grpcpage.JoinService,
	createUser *grpcuser.CreateService,
	loginUser *grpcuser.LoginService,
	getUser *grpcuser.GetService,
) *Server {
	s := &Server{}
	s.page = struct {
		create     *grpcpage.CreateService
		get        *grpcpage.GetService
		list       *grpcpage.ListService
		edit       *grpcpage.EditService
		delete     *grpcpage.DeleteService
		linkAdd    *grpcpage.LinkAddService
		linkRemove *grpcpage.LinkRemoveService
		join       *grpcpage.JoinService
	}{
		create:     createPage,
		get:        getPage,
		list:       listPages,
		edit:       editPage,
		delete:     deletePage,
		linkAdd:    addLink,
		linkRemove: removeLink,
		join:       joinPage,
	}
	s.user = struct {
		create *grpcuser.CreateService
		login  *grpcuser.LoginService
		get    *grpcuser.GetService
	}{
		create: createUser,
		login:  loginUser,
		get:    getUser,
	}
	return s
}

func (s *Server) WithUserCache(c cache.Cache[*duser.User]) {
	if s == nil {
		return
	}
	if s.user.create != nil {
		s.user.create.SetCache(c)
	}
}

func (s *Server) CreatePage(ctx context.Context, req *tsudzuriv1.CreatePageRequest) (*emptypb.Empty, error) {
	return errcode.WrapGRPC(s.page.create.Create(ctx, req))
}

func (s *Server) GetPage(ctx context.Context, req *tsudzuriv1.GetPageRequest) (*tsudzuriv1.Page, error) {
	return errcode.WrapGRPC(s.page.get.Get(ctx, req))
}

func (s *Server) ListPages(ctx context.Context, req *tsudzuriv1.ListPagesRequest) (*tsudzuriv1.ListPagesResponse, error) {
	return errcode.WrapGRPC(s.page.list.List(ctx, req))
}

func (s *Server) EditPage(ctx context.Context, req *tsudzuriv1.EditPageRequest) (*emptypb.Empty, error) {
	return errcode.WrapGRPC(s.page.edit.Edit(ctx, req))
}

func (s *Server) DeletePage(ctx context.Context, req *tsudzuriv1.DeletePageRequest) (*emptypb.Empty, error) {
	return errcode.WrapGRPC(s.page.delete.Delete(ctx, req))
}

func (s *Server) AddLink(ctx context.Context, req *tsudzuriv1.AddLinkRequest) (*emptypb.Empty, error) {
	return errcode.WrapGRPC(s.page.linkAdd.Add(ctx, req))
}

func (s *Server) RemoveLink(ctx context.Context, req *tsudzuriv1.RemoveLinkRequest) (*emptypb.Empty, error) {
	return errcode.WrapGRPC(s.page.linkRemove.Remove(ctx, req))
}

func (s *Server) JoinPage(ctx context.Context, req *tsudzuriv1.JoinPageRequest) (*emptypb.Empty, error) {
	return errcode.WrapGRPC(s.page.join.Join(ctx, req))
}

func (s *Server) CreateUser(ctx context.Context, req *tsudzuriv1.CreateUserRequest) (*tsudzuriv1.User, error) {
	return errcode.WrapGRPC(s.user.create.Create(ctx, req))
}

func (s *Server) Login(ctx context.Context, req *tsudzuriv1.LoginRequest) (*emptypb.Empty, error) {
	return errcode.WrapGRPC(s.user.login.Login(ctx, req))
}

func (s *Server) Get(ctx context.Context, req *emptypb.Empty) (*tsudzuriv1.User, error) {
	return errcode.WrapGRPC(s.user.get.Get(ctx, req))
}
