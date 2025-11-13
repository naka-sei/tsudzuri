package page

import (
	"context"

	tsudzuriv1 "github.com/naka-sei/tsudzuri/api/tsudzuri/v1"
	duser "github.com/naka-sei/tsudzuri/domain/user"
	ctxuser "github.com/naka-sei/tsudzuri/pkg/ctx/user"
	"github.com/naka-sei/tsudzuri/pkg/log"
	"github.com/naka-sei/tsudzuri/pkg/trace"
	upage "github.com/naka-sei/tsudzuri/usecase/page"
	"google.golang.org/protobuf/types/known/emptypb"
)

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

func (s *CreateService) Create(ctx context.Context, req *tsudzuriv1.CreatePageRequest) (*emptypb.Empty, error) {
	ctx, end := trace.StartSpan(ctx, "presentation/grpc/page.Create")
	defer end()

	logger := log.LoggerFromContext(ctx)
	user, ok := ctxuser.UserFromContext(ctx)
	if !ok {
		return nil, duser.ErrUserNotFound
	}

	logger.Sugar().Infof("Page create request: title=%s user_uid=%s", req.GetTitle(), user.UID())

	if _, err := s.usecase.create.Create(ctx, req.GetTitle()); err != nil {
		return nil, err
	}

	logger.Sugar().Infof("Page created: title=%s user_uid=%s", req.GetTitle(), user.UID())
	return &emptypb.Empty{}, nil
}
