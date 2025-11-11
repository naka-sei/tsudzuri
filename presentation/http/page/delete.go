package page

import (
	"context"

	duser "github.com/naka-sei/tsudzuri/domain/user"
	ctxuser "github.com/naka-sei/tsudzuri/pkg/ctx/user"
	"github.com/naka-sei/tsudzuri/pkg/log"
	"github.com/naka-sei/tsudzuri/pkg/trace"
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

func (s *DeleteService) Delete(ctx context.Context, req DeleteRequest) error {
	ctx, end := trace.StartSpan(ctx, "presentation/http/page.Delete")
	defer end()

	l := log.LoggerFromContext(ctx)
	u, ok := ctxuser.UserFromContext(ctx)
	if !ok {
		return duser.ErrUserNotFound
	}
	uid := u.UID()
	l.Sugar().Infof("Page delete request id=%s user_uid=%s", req.PageID, uid)

	if err := s.usecase.delete.Delete(ctx, req.PageID); err != nil {
		return err
	}

	l.Sugar().Infof("Page deleted: id=%s user_uid=%s", req.PageID, uid)

	return nil
}
