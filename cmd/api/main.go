package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	otelgrpc "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"

	tsudzuriv1 "github.com/naka-sei/tsudzuri/api/tsudzuri/v1"
	"github.com/naka-sei/tsudzuri/config"
	domainuser "github.com/naka-sei/tsudzuri/domain/user"
	"github.com/naka-sei/tsudzuri/infrastructure/api/firebase"
	"github.com/naka-sei/tsudzuri/infrastructure/db/postgres"
	userrepo "github.com/naka-sei/tsudzuri/infrastructure/db/user"
	"github.com/naka-sei/tsudzuri/pkg/cache"
	authinterceptor "github.com/naka-sei/tsudzuri/pkg/grpc/interceptor/auth"
	loggerinterceptor "github.com/naka-sei/tsudzuri/pkg/grpc/interceptor/logger"
	gmiddleware "github.com/naka-sei/tsudzuri/pkg/grpc/middleware"
	applog "github.com/naka-sei/tsudzuri/pkg/log"
	presentationgrpc "github.com/naka-sei/tsudzuri/presentation/grpc"
)

const (
	serviceName     = "tsudzuri-api"
	databaseSchema  = "tsudzuri"
	serverTimeout   = 10 * time.Second
	shutdownTimeout = 5 * time.Second
	userCacheTTL    = 5 * time.Minute
)

func main() {
	rootCtx := context.Background()

	conf, err := config.Load(rootCtx)
	if err != nil {
		panic(fmt.Errorf("failed to load config: %w", err))
	}

	logger, err := applog.SetupLogger(serviceName, conf.IsDebugMode)
	if err != nil {
		panic(fmt.Errorf("failed to set up logger: %w", err))
	}
	defer func() {
		_ = logger.Sync()
	}()

	sugar := logger.Sugar()

	tp := InitializeTracer(conf)
	if tp != nil {
		defer func() {
			shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
			defer cancel()
			if err := tp.Shutdown(shutdownCtx); err != nil {
				sugar.Warnf("failed to shutdown tracer provider: %v", err)
			}
		}()
	}

	connOpts := []postgres.ConnectionOption{}
	if conf.IsDebugMode {
		connOpts = append(connOpts, postgres.WithDebug())
	}

	conn, err := postgres.NewConnection(conf.TsudzuriDatabaseDSN, conf.TsudzuriDatabaseDSN, databaseSchema, connOpts...)
	if err != nil {
		sugar.Fatalf("failed to create database connection: %v", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			sugar.Warnf("failed to close database connection: %v", err)
		}
	}()

	server, err := InitializePresentationServer(conn)
	if err != nil {
		sugar.Fatalf("failed to initialize presentation server: %v", err)
	}

	authenticator, err := firebase.NewClient(conf)
	if err != nil {
		sugar.Fatalf("failed to create firebase client: %v", err)
	}
	userRepo := userrepo.NewUserRepository(conn)
	userCache := cache.NewMemoryCache[*domainuser.User](userCacheTTL)

	server.WithUserCache(userCache)

	grpcAddr := fmt.Sprintf(":%d", conf.GRPCPort)
	grpcServer, grpcListener, err := buildGRPCServer(grpcAddr, logger, conf, server, authenticator, userRepo, userCache)
	if err != nil {
		sugar.Fatalf("failed to set up gRPC server: %v", err)
	}
	defer func() {
		if err := grpcListener.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
			sugar.Warnf("failed to close gRPC listener: %v", err)
		}
	}()

	gatewayCtx, gatewayCancel := context.WithCancel(context.Background())
	defer gatewayCancel()

	mux, err := buildGatewayMux(gatewayCtx, conf)
	if err != nil {
		sugar.Fatalf("failed to build HTTP gateway: %v", err)
	}

	httpServer := buildHTTPServer(conf, mux)

	signalCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := runServers(signalCtx, sugar, grpcAddr, grpcServer, grpcListener, httpServer, gatewayCancel); err != nil {
		sugar.Fatalf("server error: %v", err)
	}

	sugar.Info("servers stopped")
}

func buildGRPCServer(
	addr string,
	logger *zap.Logger,
	conf *config.Config,
	server *presentationgrpc.Server,
	authenticator firebase.Authenticator,
	userRepo domainuser.UserRepository,
	userCache cache.Cache[*domainuser.User],
) (*grpc.Server, net.Listener, error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to listen on gRPC address: %w", err)
	}

	grpcServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			loggerinterceptor.NewLoggerUnaryServerInterceptor(logger, conf.GoogleCloudProject),
			authinterceptor.NewAuthenticationUnaryServerInterceptor(authenticator, userRepo, userCache),
		),
	)
	tsudzuriv1.RegisterTsudzuriServiceServer(grpcServer, server)

	return grpcServer, listener, nil
}

func buildGatewayMux(ctx context.Context, conf *config.Config) (*runtime.ServeMux, error) {
	opts := []runtime.ServeMuxOption{
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.HTTPBodyMarshaler{
			Marshaler: &runtime.JSONPb{
				MarshalOptions: protojson.MarshalOptions{
					UseProtoNames:   true,
					EmitUnpopulated: false,
				},
				UnmarshalOptions: protojson.UnmarshalOptions{
					DiscardUnknown: true,
				},
			},
		}),
		runtime.WithIncomingHeaderMatcher(gmiddleware.NewHeaderMatcher()),
		runtime.WithErrorHandler(gmiddleware.NewErrorHandler()),
	}
	mux := runtime.NewServeMux(opts...)

	grpcEndpoint := fmt.Sprintf("localhost:%d", conf.GRPCPort)
	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	}

	if err := tsudzuriv1.RegisterTsudzuriServiceHandlerFromEndpoint(ctx, mux, grpcEndpoint, dialOpts); err != nil {
		return nil, fmt.Errorf("failed to register tsudzuri service gateway: %w", err)
	}

	return mux, nil
}

func buildHTTPServer(conf *config.Config, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:         fmt.Sprintf(":%d", conf.Port),
		Handler:      otelhttp.NewHandler(handler, "tsudzuri-http-gateway"),
		ReadTimeout:  serverTimeout,
		WriteTimeout: serverTimeout,
		IdleTimeout:  30 * time.Second,
	}
}

func runServers(
	ctx context.Context,
	sugar *zap.SugaredLogger,
	grpcAddr string,
	grpcServer *grpc.Server,
	grpcListener net.Listener,
	httpServer *http.Server,
	gatewayCancel context.CancelFunc,
) error {
	group, groupCtx := errgroup.WithContext(ctx)

	runGRPCServer(group, sugar, grpcAddr, grpcServer, grpcListener)
	runHTTPServer(group, sugar, httpServer)
	group.Go(func() error {
		<-groupCtx.Done()
		shutdownServers(sugar, grpcServer, httpServer, gatewayCancel)
		return nil
	})

	return group.Wait()
}

func runGRPCServer(group *errgroup.Group, sugar *zap.SugaredLogger, addr string, server *grpc.Server, listener net.Listener) {
	group.Go(func() error {
		sugar.Infow("gRPC server starting", zap.String("addr", addr))
		if err := server.Serve(listener); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			return err
		}
		return nil
	})
}

func runHTTPServer(group *errgroup.Group, sugar *zap.SugaredLogger, server *http.Server) {
	group.Go(func() error {
		sugar.Infow("HTTP gateway server starting", zap.String("addr", server.Addr))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	})
}

func shutdownServers(
	sugar *zap.SugaredLogger,
	grpcServer *grpc.Server,
	httpServer *http.Server,
	gatewayCancel context.CancelFunc,
) {
	sugar.Info("shutdown initiated")
	grpcServer.GracefulStop()
	gatewayCancel()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		sugar.Errorf("HTTP server shutdown error: %v", err)
	}
}
