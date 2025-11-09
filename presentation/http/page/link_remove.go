package page

import (
	"context"

	"github.com/naka-sei/tsudzuri/pkg/log"
	"github.com/naka-sei/tsudzuri/pkg/trace"
	"github.com/naka-sei/tsudzuri/presentation/http/response"
	upage "github.com/naka-sei/tsudzuri/usecase/page"
)

type LinkRemoveRequest struct {
	PageID string `json:"page_id"`
	URL    string `json:"url"`
}

type LinkRemoveService struct {
	usecase struct{ remove upage.LinkRemoveUseCase }
}

func NewLinkRemoveService(lr upage.LinkRemoveUseCase) *LinkRemoveService {
	return &LinkRemoveService{usecase: struct{ remove upage.LinkRemoveUseCase }{remove: lr}}
}

func (s *LinkRemoveService) LinkRemove(ctx context.Context, req LinkRemoveRequest) (response.EmptyResponse, error) {
	ctx, end := trace.StartSpan(ctx, "presentation/http/page.LinkRemove")
	defer end()

	l := log.LoggerFromContext(ctx)
	l.Sugar().Infof("Page link remove request page_id=%s url=%s", req.PageID, req.URL)

	input := upage.LinkRemoveUsecaseInput{PageID: req.PageID, URL: req.URL}
	if err := s.usecase.remove.LinkRemove(ctx, input); err != nil {
		return response.EmptyResponse{}, err
	}

	return response.EmptyResponse{}, nil
}
