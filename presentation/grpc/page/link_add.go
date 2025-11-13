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

type LinkAddService struct {
	usecase struct {
		linkAdd upage.LinkAddUseCase
	}
}

func NewLinkAddService(lu upage.LinkAddUseCase) *LinkAddService {
	return &LinkAddService{
		usecase: struct{ linkAdd upage.LinkAddUseCase }{linkAdd: lu},
	}
}

func (s *LinkAddService) Add(ctx context.Context, req *tsudzuriv1.AddLinkRequest) (*emptypb.Empty, error) {
	ctx, end := trace.StartSpan(ctx, "presentation/grpc/page.LinkAdd")
	defer end()

	logger := log.LoggerFromContext(ctx)
	user, ok := ctxuser.UserFromContext(ctx)
	if !ok {
		return nil, duser.ErrUserNotFound
	}

	logger.Sugar().Infof("Page link add request page_id=%s url=%s user_uid=%s", req.GetPageId(), req.GetUrl(), user.UID())

	input := upage.LinkAddUsecaseInput{
		PageID: req.GetPageId(),
		URL:    req.GetUrl(),
		Memo:   req.GetMemo(),
	}

	if err := s.usecase.linkAdd.LinkAdd(ctx, input); err != nil {
		return nil, err
	}

	logger.Sugar().Infof("Page link add succeeded page_id=%s user_uid=%s", req.GetPageId(), user.UID())
	return &emptypb.Empty{}, nil
}
