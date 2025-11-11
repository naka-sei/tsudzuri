package page

import (
	"context"

	duser "github.com/naka-sei/tsudzuri/domain/user"
	ctxuser "github.com/naka-sei/tsudzuri/pkg/ctx/user"
	"github.com/naka-sei/tsudzuri/pkg/log"
	"github.com/naka-sei/tsudzuri/pkg/trace"
	upage "github.com/naka-sei/tsudzuri/usecase/page"
)

type ListRequest struct{}

type LinkResponse struct {
	URL      string `json:"url"`
	Memo     string `json:"memo"`
	Priority int    `json:"priority"`
}

type PageResponse struct {
	ID         string         `json:"id"`
	Title      string         `json:"title"`
	InviteCode string         `json:"invite_code"`
	Links      []LinkResponse `json:"links"`
}

type ListService struct {
	usecase struct {
		list upage.ListUsecase
	}
}

func NewListService(lu upage.ListUsecase) *ListService {
	return &ListService{usecase: struct{ list upage.ListUsecase }{list: lu}}
}

func (s *ListService) List(ctx context.Context, req ListRequest) ([]PageResponse, error) {
	ctx, end := trace.StartSpan(ctx, "presentation/http/page.List")
	defer end()

	l := log.LoggerFromContext(ctx)
	u, ok := ctxuser.UserFromContext(ctx)
	if !ok {
		return nil, duser.ErrUserNotFound
	}
	uid := u.UID()
	l.Sugar().Infof("Page list request user_uid=%s", uid)

	pages, err := s.usecase.list.List(ctx)
	if err != nil {
		return nil, err
	}

	if len(pages) == 0 {
		l.Sugar().Infof("Page list responded: count=0 user_uid=%s", uid)
		return nil, nil
	}

	res := make([]PageResponse, 0, len(pages))
	for _, p := range pages {
		pr := PageResponse{
			ID:         p.ID(),
			Title:      p.Title(),
			InviteCode: p.InviteCode(),
		}
		links := p.Links()
		if len(links) > 0 {
			prs := make([]LinkResponse, 0, len(links))
			for _, lnk := range links {
				prs = append(prs, LinkResponse{URL: lnk.URL(), Memo: lnk.Memo(), Priority: lnk.Priority()})
			}
			pr.Links = prs
		}
		res = append(res, pr)
	}

	l.Sugar().Infof("Page list responded: count=%d user_uid=%s", len(res), uid)
	return res, nil
}
