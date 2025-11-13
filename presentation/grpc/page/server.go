package page

import (
	"context"

	tsudzuriv1 "github.com/naka-sei/tsudzuri/api/tsudzuri/v1"
	"github.com/naka-sei/tsudzuri/presentation/errcode"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Server struct {
	tsudzuriv1.UnimplementedPageServiceServer

	create     *CreateService
	get        *GetService
	list       *ListService
	edit       *EditService
	delete     *DeleteService
	linkAdd    *LinkAddService
	linkRemove *LinkRemoveService
}

func NewServer(
	create *CreateService,
	get *GetService,
	list *ListService,
	edit *EditService,
	delete *DeleteService,
	linkAdd *LinkAddService,
	linkRemove *LinkRemoveService,
) *Server {
	return &Server{
		create:     create,
		get:        get,
		list:       list,
		edit:       edit,
		delete:     delete,
		linkAdd:    linkAdd,
		linkRemove: linkRemove,
	}
}

func (s *Server) CreatePage(ctx context.Context, req *tsudzuriv1.CreatePageRequest) (*emptypb.Empty, error) {
	return errcode.WrapGRPC(s.create.Create(ctx, req))
}

func (s *Server) GetPage(ctx context.Context, req *tsudzuriv1.GetPageRequest) (*tsudzuriv1.Page, error) {
	return errcode.WrapGRPC(s.get.Get(ctx, req))
}

func (s *Server) ListPages(ctx context.Context, req *tsudzuriv1.ListPagesRequest) (*tsudzuriv1.ListPagesResponse, error) {
	return errcode.WrapGRPC(s.list.List(ctx, req))
}

func (s *Server) EditPage(ctx context.Context, req *tsudzuriv1.EditPageRequest) (*emptypb.Empty, error) {
	return errcode.WrapGRPC(s.edit.Edit(ctx, req))
}

func (s *Server) DeletePage(ctx context.Context, req *tsudzuriv1.DeletePageRequest) (*emptypb.Empty, error) {
	return errcode.WrapGRPC(s.delete.Delete(ctx, req))
}

func (s *Server) AddLink(ctx context.Context, req *tsudzuriv1.AddLinkRequest) (*emptypb.Empty, error) {
	return errcode.WrapGRPC(s.linkAdd.Add(ctx, req))
}

func (s *Server) RemoveLink(ctx context.Context, req *tsudzuriv1.RemoveLinkRequest) (*emptypb.Empty, error) {
	return errcode.WrapGRPC(s.linkRemove.Remove(ctx, req))
}
