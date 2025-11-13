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
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	tsudzuriv1 "github.com/naka-sei/tsudzuri/api/tsudzuri/v1"
	"github.com/naka-sei/tsudzuri/config"
	domainuser "github.com/naka-sei/tsudzuri/domain/user"
	"github.com/naka-sei/tsudzuri/infrastructure/api/firebase"
	"github.com/naka-sei/tsudzuri/infrastructure/db/postgres"
	userrepo "github.com/naka-sei/tsudzuri/infrastructure/db/user"
	"github.com/naka-sei/tsudzuri/pkg/cache"
	authinterceptor "github.com/naka-sei/tsudzuri/pkg/grpc/interceptor/auth"
	loggerinterceptor "github.com/naka-sei/tsudzuri/pkg/grpc/interceptor/logger"
	applog "github.com/naka-sei/tsudzuri/pkg/log"
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
	grpcListener, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		sugar.Fatalf("failed to listen on gRPC address: %v", err)
	}

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			loggerinterceptor.NewLoggerUnaryServerInterceptor(logger, conf.GoogleCloudProject),
			authinterceptor.NewAuthenticationUnaryServerInterceptor(authenticator, userRepo, userCache),
		),
	)
	server.RegisterGRPC(grpcServer)

	signalCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	group, groupCtx := errgroup.WithContext(signalCtx)

	group.Go(func() error {
		sugar.Infow("gRPC server starting", zap.String("addr", grpcAddr))
		if err := grpcServer.Serve(grpcListener); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			return err
		}
		return nil
	})

	gatewayCtx, gatewayCancel := context.WithCancel(context.Background())
	defer gatewayCancel()

	grpcEndpoint := fmt.Sprintf("localhost:%d", conf.GRPCPort)
	dialOpts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	mux := runtime.NewServeMux()
	if err := tsudzuriv1.RegisterPageServiceHandlerFromEndpoint(gatewayCtx, mux, grpcEndpoint, dialOpts); err != nil {
		sugar.Fatalf("failed to register page service gateway: %v", err)
	}
	if err := tsudzuriv1.RegisterUserServiceHandlerFromEndpoint(gatewayCtx, mux, grpcEndpoint, dialOpts); err != nil {
		sugar.Fatalf("failed to register user service gateway: %v", err)
	}

	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", conf.Port),
		Handler:      mux,
		ReadTimeout:  serverTimeout,
		WriteTimeout: serverTimeout,
		IdleTimeout:  30 * time.Second,
	}

	group.Go(func() error {
		sugar.Infow("HTTP gateway server starting", zap.String("addr", httpServer.Addr))
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	})

	group.Go(func() error {
		<-groupCtx.Done()
		sugar.Info("shutdown initiated")
		grpcServer.GracefulStop()
		gatewayCancel()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			sugar.Errorf("HTTP server shutdown error: %v", err)
		}
		return nil
	})

	if err := group.Wait(); err != nil {
		sugar.Fatalf("server error: %v", err)
	}

	sugar.Info("servers stopped")
}
