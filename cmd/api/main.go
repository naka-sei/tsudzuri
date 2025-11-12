package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/naka-sei/tsudzuri/config"
	domainuser "github.com/naka-sei/tsudzuri/domain/user"
	"github.com/naka-sei/tsudzuri/infrastructure/api/firebase"
	chiadapter "github.com/naka-sei/tsudzuri/infrastructure/api/http/chi"
	"github.com/naka-sei/tsudzuri/infrastructure/db/postgres"
	userrepo "github.com/naka-sei/tsudzuri/infrastructure/db/user"
	"github.com/naka-sei/tsudzuri/pkg/cache"
	"github.com/naka-sei/tsudzuri/pkg/http/middleware/authentication"
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
	ctx := context.Background()

	conf, err := config.Load(ctx)
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

	router := chi.NewRouter()
	router.Use(authentication.AuthHTTPMiddleware(authenticator, userRepo, userCache))
	server.Route(chiadapter.New(router))

	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", conf.Port),
		Handler:      router,
		ReadTimeout:  serverTimeout,
		WriteTimeout: serverTimeout,
		IdleTimeout:  30 * time.Second,
	}

	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-shutdownCh
		sugar.Info("Shutting down HTTP server")
		ctx, cancel := context.WithTimeout(context.Background(), serverTimeout)
		defer cancel()
		if err := httpServer.Shutdown(ctx); err != nil {
			sugar.Errorf("HTTP server shutdown error: %v", err)
		}
	}()

	sugar.Infow("server starting", zap.String("addr", httpServer.Addr))
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		sugar.Fatalf("HTTP server error: %v", err)
	}
	sugar.Info("HTTP server stopped")
}
