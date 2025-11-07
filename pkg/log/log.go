package log

import (
	"context"
	"fmt"

	"github.com/blendle/zapdriver"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// SetupLogger sets up the logger.
func SetupLogger(serviceName string, debugMode bool) (logger *zap.Logger, err error) {
	if debugMode {
		return newDevelopmentCLogger(serviceName, "debug")
	}
	return newProductionCLogger(serviceName, "info")
}

// newProductionCLogger creates a new production logger.
// The level is a logging level string (e.g., "debug", "info", "warn", "error").
func newProductionCLogger(serviceName, level string, options ...zap.Option) (*zap.Logger, error) {
	cfg := zapdriver.NewProductionConfig()
	err := cfg.Level.UnmarshalText([]byte(level))
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal log level")
	}
	cfg.Sampling = nil
	cfg.OutputPaths = []string{"stdout"}
	cfg.ErrorOutputPaths = []string{"stderr"}
	return cfg.Build(zapOptions(serviceName, options...)...)
}

// newDevelopmentCLogger creates a new development logger.
// The level is a logging level string (e.g., "debug", "info", "warn", "error").
func newDevelopmentCLogger(serviceName, level string, options ...zap.Option) (*zap.Logger, error) {
	cfg := zapdriver.NewDevelopmentConfig()
	err := cfg.Level.UnmarshalText([]byte(level))
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal log level")
	}
	cfg.Sampling = nil
	cfg.OutputPaths = []string{"stdout"}
	cfg.ErrorOutputPaths = []string{"stderr"}
	return cfg.Build(zapOptions(serviceName, options...)...)
}

func zapOptions(serviceName string, options ...zap.Option) []zap.Option {
	return append(options, zapdriver.WrapCore(
		zapdriver.ServiceName(serviceName),
	))
}

const LogTypeLabel = "log_type"

type ctxKey struct{}

type loggerContext struct {
	logger    *zap.Logger
	projectID string
}

// NewLoggerContext creates a new logger context.
func NewLoggerContext(ctx context.Context, logger *zap.Logger, projectID string) context.Context {
	lc := &loggerContext{
		logger:    logger,
		projectID: projectID,
	}
	return context.WithValue(ctx, ctxKey{}, lc)
}

// LoggerFromContext returns a logger with the given log type.
func LoggerFromContext(ctx context.Context) *zap.Logger {
	l, ok := ctx.Value(ctxKey{}).(*loggerContext)
	if !ok {
		return zap.NewNop()
	}
	return l.loggerWithTraceInfo(ctx).With(zapdriver.Label(LogTypeLabel, "tsudzuri_app"))
}

// loggerWithTraceInfo returns a logger with trace info from context.
func (l *loggerContext) loggerWithTraceInfo(ctx context.Context) *zap.Logger {
	sc := trace.SpanContextFromContext(ctx)
	return l.logger.With(
		zap.String("logging.googleapis.com/trace", fmt.Sprintf("projects/%s/traces/%s", l.projectID, sc.TraceID().String())),
		zap.String("logging.googleapis.com/spanId", sc.SpanID().String()),
		zap.String("logging.googleapis.com/trace_sampled", string(sc.TraceFlags())),
	)
}
