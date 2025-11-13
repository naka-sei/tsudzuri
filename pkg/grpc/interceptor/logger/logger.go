package interceptor

import (
	"context"

	"github.com/naka-sei/tsudzuri/pkg/log"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func NewLoggerUnaryServerInterceptor(logger *zap.Logger, projectID string) grpc.UnaryServerInterceptor {
	if logger == nil {
		logger = zap.NewNop()
	}

	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		requestLogger := logger.With(zap.String("grpc.method", info.FullMethod))
		ctx = log.NewLoggerContext(ctx, requestLogger, projectID)
		return handler(ctx, req)
	}
}
