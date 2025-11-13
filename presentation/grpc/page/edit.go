package page

import (
	"context"

	tsudzuriv1 "github.com/naka-sei/tsudzuri/api/tsudzuri/v1"
	dpage "github.com/naka-sei/tsudzuri/domain/page"
	duser "github.com/naka-sei/tsudzuri/domain/user"
	ctxuser "github.com/naka-sei/tsudzuri/pkg/ctx/user"
	"github.com/naka-sei/tsudzuri/pkg/log"
	"github.com/naka-sei/tsudzuri/pkg/trace"
	upage "github.com/naka-sei/tsudzuri/usecase/page"
	"google.golang.org/protobuf/types/known/emptypb"
)

type EditService struct {
	usecase struct {
		edit upage.EditUsecase
	}
}

func NewEditService(eu upage.EditUsecase) *EditService {
	return &EditService{
		usecase: struct{ edit upage.EditUsecase }{edit: eu},
	}
}

func (s *EditService) Edit(ctx context.Context, req *tsudzuriv1.EditPageRequest) (*emptypb.Empty, error) {
	ctx, end := trace.StartSpan(ctx, "presentation/grpc/page.Edit")
	defer end()

	logger := log.LoggerFromContext(ctx)
	user, ok := ctxuser.UserFromContext(ctx)
	if !ok {
		return nil, duser.ErrUserNotFound
	}

	logger.Sugar().Infof("Page edit request page_id=%s title=%s links=%d user_uid=%s", req.GetPageId(), req.GetTitle(), len(req.GetLinks()), user.UID())

	links := make(dpage.Links, 0, len(req.GetLinks()))
	for _, lnk := range req.GetLinks() {
		links = append(links, dpage.ReconstructLink(lnk.GetUrl(), lnk.GetMemo(), int(lnk.GetPriority())))
	}

	if err := s.usecase.edit.Edit(ctx, req.GetPageId(), req.GetTitle(), links); err != nil {
		return nil, err
	}

	logger.Sugar().Infof("Page edit succeeded page_id=%s user_uid=%s", req.GetPageId(), user.UID())
	return &emptypb.Empty{}, nil
}
