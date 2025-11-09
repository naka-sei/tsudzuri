package page

import (
	"context"

	"github.com/naka-sei/tsudzuri/pkg/log"
	"github.com/naka-sei/tsudzuri/pkg/trace"
	"github.com/naka-sei/tsudzuri/presentation/http/response"
	upage "github.com/naka-sei/tsudzuri/usecase/page"
)

type GetRequest struct {
	PageID string `path:"id"`
}

type GetService struct {
	usecase struct{ get upage.GetUsecase }
}

func NewGetService(gu upage.GetUsecase) *GetService {
	return &GetService{usecase: struct{ get upage.GetUsecase }{get: gu}}
}

func (s *GetService) Get(ctx context.Context, req GetRequest) (PageResponse, error) {
	ctx, end := trace.StartSpan(ctx, "presentation/http/page.Get")
	defer end()

	l := log.LoggerFromContext(ctx)
	l.Sugar().Infof("Page get request id=%s", req.PageID)

	p, err := s.usecase.get.Get(ctx, req.PageID)
	if err != nil {
		return PageResponse{}, err
	}

	res := PageResponse{
		ID:         p.ID(),
		Title:      p.Title(),
		InviteCode: p.InviteCode(),
	}

	if ls := p.Links(); len(ls) > 0 {
		prs := make([]LinkResponse, 0, len(ls))
		for _, lnk := range ls {
			prs = append(prs, LinkResponse{URL: lnk.URL(), Memo: lnk.Memo(), Priority: lnk.Priority()})
		}
		res.Links = prs
	}

	_ = response.EmptyResponse{} // keep response package available for consistency with other handlers
	return res, nil
}
