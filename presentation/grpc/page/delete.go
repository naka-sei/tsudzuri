package page

import (
	"context"

	tsudzuriv1 "github.com/naka-sei/tsudzuri/api/tsudzuri/v1"
	duser "github.com/naka-sei/tsudzuri/domain/user"
	ctxuser "github.com/naka-sei/tsudzuri/pkg/ctx/user"
	"github.com/naka-sei/tsudzuri/pkg/log"
	"github.com/naka-sei/tsudzuri/pkg/trace"
	upage "github.com/naka-sei/tsudzuri/usecase/page"
	"google.golang.org/protobuf/types/known/emptypb"
)

type DeleteService struct {
	usecase struct {
		delete upage.DeleteUsecase
	}
}

func NewDeleteService(du upage.DeleteUsecase) *DeleteService {
	return &DeleteService{
		usecase: struct{ delete upage.DeleteUsecase }{delete: du},
	}
}

func (s *DeleteService) Delete(ctx context.Context, req *tsudzuriv1.DeletePageRequest) (*emptypb.Empty, error) {
	ctx, end := trace.StartSpan(ctx, "presentation/grpc/page.Delete")
	defer end()

	logger := log.LoggerFromContext(ctx)
	user, ok := ctxuser.UserFromContext(ctx)
	if !ok {
		return nil, duser.ErrUserNotFound
	}

	logger.Sugar().Infof("Page delete request page_id=%s user_uid=%s", req.GetPageId(), user.UID())

	if err := s.usecase.delete.Delete(ctx, req.GetPageId()); err != nil {
		return nil, err
	}

	logger.Sugar().Infof("Page delete succeeded page_id=%s user_uid=%s", req.GetPageId(), user.UID())
	return &emptypb.Empty{}, nil
}
