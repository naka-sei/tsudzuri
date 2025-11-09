package page

import (
	"context"

	"github.com/naka-sei/tsudzuri/pkg/log"
	"github.com/naka-sei/tsudzuri/pkg/trace"
	"github.com/naka-sei/tsudzuri/presentation/http/response"
	upage "github.com/naka-sei/tsudzuri/usecase/page"
)

type LinkAddRequest struct {
	PageID string `path:"id"`
	URL    string `json:"url"`
	Memo   string `json:"memo"`
}

type LinkAddService struct {
	usecase struct{ add upage.LinkAddUseCase }
}

func NewLinkAddService(la upage.LinkAddUseCase) *LinkAddService {
	return &LinkAddService{usecase: struct{ add upage.LinkAddUseCase }{add: la}}
}

func (s *LinkAddService) LinkAdd(ctx context.Context, req LinkAddRequest) (response.EmptyResponse, error) {
	ctx, end := trace.StartSpan(ctx, "presentation/http/page.LinkAdd")
	defer end()

	l := log.LoggerFromContext(ctx)
	l.Sugar().Infof("Page link add request page_id=%s url=%s", req.PageID, req.URL)

	input := upage.LinkAddUsecaseInput{PageID: req.PageID, URL: req.URL, Memo: req.Memo}
	if err := s.usecase.add.LinkAdd(ctx, input); err != nil {
		return response.EmptyResponse{}, err
	}

	return response.EmptyResponse{}, nil
}
