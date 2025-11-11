package page

import (
	"context"

	duser "github.com/naka-sei/tsudzuri/domain/user"
	ctxuser "github.com/naka-sei/tsudzuri/pkg/ctx/user"
	"github.com/naka-sei/tsudzuri/pkg/log"
	"github.com/naka-sei/tsudzuri/pkg/trace"
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

func (s *GetService) Get(ctx context.Context, req GetRequest) (*PageResponse, error) {
	ctx, end := trace.StartSpan(ctx, "presentation/http/page.Get")
	defer end()

	l := log.LoggerFromContext(ctx)
	u, ok := ctxuser.UserFromContext(ctx)
	if !ok {
		return nil, duser.ErrUserNotFound
	}
	uid := u.UID()
	l.Sugar().Infof("Page get request id=%s user_uid=%s", req.PageID, uid)

	p, err := s.usecase.get.Get(ctx, req.PageID)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, nil
	}

	res := &PageResponse{
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

	l.Sugar().Infof("Page get responded: id=%s links=%d user_uid=%s", p.ID(), len(res.Links), uid)

	return res, nil
}
