package page

import (
	"context"

	duser "github.com/naka-sei/tsudzuri/domain/user"
	ctxuser "github.com/naka-sei/tsudzuri/pkg/ctx/user"
	"github.com/naka-sei/tsudzuri/pkg/log"
	"github.com/naka-sei/tsudzuri/pkg/trace"
	upage "github.com/naka-sei/tsudzuri/usecase/page"
)

type CreateRequest struct {
	Title string `json:"title"`
}

type CreateService struct {
	usecase struct {
		create upage.CreateUsecase
	}
}

func NewCreateService(cu upage.CreateUsecase) *CreateService {
	return &CreateService{
		usecase: struct{ create upage.CreateUsecase }{create: cu},
	}
}

// Create is a transport-agnostic presentation handler.
// It expects a context and a request DTO; returns a response DTO or error.
func (s *CreateService) Create(ctx context.Context, req CreateRequest) error {
	ctx, end := trace.StartSpan(ctx, "presentation/http/page.Create")
	defer end()

	u, ok := ctxuser.UserFromContext(ctx)
	if !ok {
		return duser.ErrUserNotFound
	}

	l := log.LoggerFromContext(ctx)
	l.Sugar().Infof("Page create request: title=%s user_uid=%s", req.Title, u.UID())

	p, err := s.usecase.create.Create(ctx, req.Title)
	if err != nil {
		return err
	}

	l.Sugar().Infof("Page created: id=%s title=%s user_uid=%s", p.ID(), p.Title(), u.UID())
	return nil
}
