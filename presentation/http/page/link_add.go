package page

import (
	"context"

	duser "github.com/naka-sei/tsudzuri/domain/user"
	ctxuser "github.com/naka-sei/tsudzuri/pkg/ctx/user"
	"github.com/naka-sei/tsudzuri/pkg/log"
	"github.com/naka-sei/tsudzuri/pkg/trace"
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

func (s *LinkAddService) LinkAdd(ctx context.Context, req LinkAddRequest) error {
	ctx, end := trace.StartSpan(ctx, "presentation/http/page.LinkAdd")
	defer end()

	l := log.LoggerFromContext(ctx)
	u, ok := ctxuser.UserFromContext(ctx)
	if !ok {
		return duser.ErrUserNotFound
	}
	uid := u.UID()
	l.Sugar().Infof("Page link add request page_id=%s url=%s user_uid=%s", req.PageID, req.URL, uid)

	input := upage.LinkAddUsecaseInput{PageID: req.PageID, URL: req.URL, Memo: req.Memo}
	if err := s.usecase.add.LinkAdd(ctx, input); err != nil {
		return err
	}

	l.Sugar().Infof("Page link added: page_id=%s url=%s user_uid=%s", req.PageID, req.URL, uid)

	return nil
}
