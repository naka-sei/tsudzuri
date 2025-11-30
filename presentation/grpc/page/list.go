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

type ListService struct {
	usecase struct {
		list upage.ListUsecase
	}
}

func NewListService(lu upage.ListUsecase) *ListService {
	return &ListService{
		usecase: struct{ list upage.ListUsecase }{list: lu},
	}
}

func (s *ListService) List(ctx context.Context, _ *tsudzuriv1.ListPagesRequest) (*tsudzuriv1.ListPagesResponse, error) {
	ctx, end := trace.StartSpan(ctx, "presentation/grpc/page.List")
	defer end()

	logger := log.LoggerFromContext(ctx)
	user, ok := ctxuser.UserFromContext(ctx)
	if !ok {
		return nil, duser.ErrUserNotFound
	}

	logger.Sugar().Infof("Page list request user_uid=%s", user.UID())

	pages, err := s.usecase.list.List(ctx)
	if err != nil {
		return nil, err
	}

	resp := &tsudzuriv1.ListPagesResponse{}
	if len(pages) > 0 {
		resp.Pages = make([]*tsudzuriv1.Page, 0, len(pages))
		for _, p := range pages {
			resp.Pages = append(resp.Pages, toProtoPage(p, user))
		}
	}

	logger.Sugar().Infof("Page list responded: count=%d user_uid=%s", len(resp.GetPages()), user.UID())
	return resp, nil
}
