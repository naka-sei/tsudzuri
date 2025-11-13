package page

import (
	"context"

	tsudzuriv1 "github.com/naka-sei/tsudzuri/api/tsudzuri/v1"
	"github.com/naka-sei/tsudzuri/pkg/log"
	"github.com/naka-sei/tsudzuri/pkg/trace"
	upage "github.com/naka-sei/tsudzuri/usecase/page"
	"google.golang.org/protobuf/types/known/emptypb"
)

type JoinService struct {
	usecase struct {
		join upage.JoinUsecase
	}
}

func NewJoinService(ju upage.JoinUsecase) *JoinService {
	return &JoinService{usecase: struct{ join upage.JoinUsecase }{join: ju}}
}

func (s *JoinService) Join(ctx context.Context, req *tsudzuriv1.JoinPageRequest) (*emptypb.Empty, error) {
	ctx, end := trace.StartSpan(ctx, "presentation/grpc/page.Join")
	defer end()

	logger := log.LoggerFromContext(ctx)
	logger.Sugar().Infof("Join page request: page_id=%s", req.GetPageId())

	if err := s.usecase.join.Join(ctx, req.GetPageId(), req.GetInviteCode()); err != nil {
		return nil, err
	}

	logger.Sugar().Infof("Join page succeeded: page_id=%s", req.GetPageId())
	return &emptypb.Empty{}, nil
}
