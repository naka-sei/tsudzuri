package page

import (
	"context"

	dpage "github.com/naka-sei/tsudzuri/domain/page"
	duser "github.com/naka-sei/tsudzuri/domain/user"
	ctxuser "github.com/naka-sei/tsudzuri/pkg/ctx/user"
	"github.com/naka-sei/tsudzuri/pkg/log"
	"github.com/naka-sei/tsudzuri/pkg/trace"
	upage "github.com/naka-sei/tsudzuri/usecase/page"
)

type EditRequest struct {
	PageID string        `path:"id"`
	Title  string        `json:"title"`
	Links  []LinkRequest `json:"links"`
}

type LinkRequest struct {
	URL      string `json:"url"`
	Memo     string `json:"memo"`
	Priority int    `json:"priority"`
}

type EditService struct {
	usecase struct{ edit upage.EditUsecase }
}

func NewEditService(eu upage.EditUsecase) *EditService {
	return &EditService{usecase: struct{ edit upage.EditUsecase }{edit: eu}}
}

func (s *EditService) Edit(ctx context.Context, req EditRequest) error {
	ctx, end := trace.StartSpan(ctx, "presentation/http/page.Edit")
	defer end()

	l := log.LoggerFromContext(ctx)
	u, ok := ctxuser.UserFromContext(ctx)
	if !ok {
		return duser.ErrUserNotFound
	}
	uid := u.UID()
	l.Sugar().Infof("Page edit request id=%s title=%s user_uid=%s", req.PageID, req.Title, uid)

	var links dpage.Links
	if len(req.Links) > 0 {
		links = make(dpage.Links, 0, len(req.Links))
		for _, li := range req.Links {
			links = append(links, dpage.ReconstructLink(li.URL, li.Memo, li.Priority))
		}
	}

	if err := s.usecase.edit.Edit(ctx, req.PageID, req.Title, links); err != nil {
		return err
	}

	l.Sugar().Infof("Page edited: id=%s title=%s user_uid=%s", req.PageID, req.Title, uid)

	return nil
}
