package page

import (
	"context"

	tsudzuriv1 "github.com/naka-sei/tsudzuri/api/tsudzuri/v1"
	duser "github.com/naka-sei/tsudzuri/domain/user"
	ctxuser "github.com/naka-sei/tsudzuri/pkg/ctx/user"
	"github.com/naka-sei/tsudzuri/pkg/log"
	"github.com/naka-sei/tsudzuri/pkg/trace"
	upage "github.com/naka-sei/tsudzuri/usecase/page"
)

type GetService struct {
	usecase struct {
		get upage.GetUsecase
	}
}

func NewGetService(gu upage.GetUsecase) *GetService {
	return &GetService{
		usecase: struct{ get upage.GetUsecase }{get: gu},
	}
}

func (s *GetService) Get(ctx context.Context, req *tsudzuriv1.GetPageRequest) (*tsudzuriv1.Page, error) {
	ctx, end := trace.StartSpan(ctx, "presentation/grpc/page.Get")
	defer end()

	logger := log.LoggerFromContext(ctx)
	user, ok := ctxuser.UserFromContext(ctx)
	if !ok {
		return nil, duser.ErrUserNotFound
	}

	logger.Sugar().Infof("Page get request page_id=%s user_uid=%s", req.GetPageId(), user.UID())

	page, err := s.usecase.get.Get(ctx, req.GetPageId())
	if err != nil {
		return nil, err
	}

	resp := toProtoPage(page)
	linksCount := 0
	if resp != nil {
		linksCount = len(resp.GetLinks())
	}
	logger.Sugar().Infof("Page get responded: page_id=%s links=%d user_uid=%s", req.GetPageId(), linksCount, user.UID())

	return resp, nil
}
