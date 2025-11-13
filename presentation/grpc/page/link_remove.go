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

type LinkRemoveService struct {
	usecase struct {
		linkRemove upage.LinkRemoveUseCase
	}
}

func NewLinkRemoveService(lu upage.LinkRemoveUseCase) *LinkRemoveService {
	return &LinkRemoveService{
		usecase: struct{ linkRemove upage.LinkRemoveUseCase }{linkRemove: lu},
	}
}

func (s *LinkRemoveService) Remove(ctx context.Context, req *tsudzuriv1.RemoveLinkRequest) (*emptypb.Empty, error) {
	ctx, end := trace.StartSpan(ctx, "presentation/grpc/page.LinkRemove")
	defer end()

	logger := log.LoggerFromContext(ctx)
	user, ok := ctxuser.UserFromContext(ctx)
	if !ok {
		return nil, duser.ErrUserNotFound
	}

	logger.Sugar().Infof("Page link remove request page_id=%s url=%s user_uid=%s", req.GetPageId(), req.GetUrl(), user.UID())

	input := upage.LinkRemoveUsecaseInput{
		PageID: req.GetPageId(),
		URL:    req.GetUrl(),
	}

	if err := s.usecase.linkRemove.LinkRemove(ctx, input); err != nil {
		return nil, err
	}

	logger.Sugar().Infof("Page link remove succeeded page_id=%s user_uid=%s", req.GetPageId(), user.UID())
	return &emptypb.Empty{}, nil
}
