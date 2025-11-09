package page

import (
	"context"

	"github.com/naka-sei/tsudzuri/pkg/log"
	"github.com/naka-sei/tsudzuri/pkg/trace"
	"github.com/naka-sei/tsudzuri/presentation/http/response"
	upage "github.com/naka-sei/tsudzuri/usecase/page"
)

type DeleteRequest struct {
	PageID string `path:"id"`
}

type DeleteService struct {
	usecase struct{ delete upage.DeleteUsecase }
}

func NewDeleteService(du upage.DeleteUsecase) *DeleteService {
	return &DeleteService{usecase: struct{ delete upage.DeleteUsecase }{delete: du}}
}

func (s *DeleteService) Delete(ctx context.Context, req DeleteRequest) (response.EmptyResponse, error) {
	ctx, end := trace.StartSpan(ctx, "presentation/http/page.Delete")
	defer end()

	l := log.LoggerFromContext(ctx)
	l.Sugar().Infof("Page delete request id=%s", req.PageID)

	if err := s.usecase.delete.Delete(ctx, req.PageID); err != nil {
		return response.EmptyResponse{}, err
	}

	return response.EmptyResponse{}, nil
}
