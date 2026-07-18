package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	nethttp "github.com/alamjad/marketplace/internal/delivery/http"
	"github.com/alamjad/marketplace/internal/delivery/http/handler"
	"github.com/alamjad/marketplace/internal/delivery/http/middleware"
	"github.com/alamjad/marketplace/internal/platform/auth"
	"github.com/alamjad/marketplace/internal/platform/config"
	"github.com/alamjad/marketplace/internal/platform/db"
	"github.com/alamjad/marketplace/internal/platform/storage"
	"github.com/alamjad/marketplace/internal/repository/postgres"
	authuc "github.com/alamjad/marketplace/internal/usecase/auth"
	"github.com/alamjad/marketplace/internal/usecase/catalog"
	listinguc "github.com/alamjad/marketplace/internal/usecase/listing"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	if err := run(); err != nil {
		logger.Error("server terminated", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()
	slog.Info("connected to database")

	// Platform services.
	tokens := auth.NewManager(cfg.JWTAccessSecret, cfg.JWTAccessTTL, cfg.JWTRefreshTTL)
	store, err := storage.NewLocal(cfg.UploadDir, cfg.PublicBaseURL)
	if err != nil {
		return err
	}
	repo := postgres.New(pool)

	// Usecases.
	authService := authuc.NewService(repo, tokens)
	catalogService := catalog.NewService(repo)
	listingService := listinguc.NewService(repo)

	// HTTP layer.
	router := nethttp.NewRouter(nethttp.RouterDeps{
		AllowedOrigins: cfg.CORSAllowedOrigins,
		UploadDir:      cfg.UploadDir,
		Auth:           middleware.NewAuthenticator(tokens),
		Health:         handler.NewHealthHandler(pool),
		AuthH:          handler.NewAuthHandler(authService),
		Catalog:        handler.NewCatalogHandler(catalogService),
		Listing:        handler.NewListingHandler(listingService, store),
	})

	srv := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		slog.Info("http server listening", "port", cfg.HTTPPort, "env", cfg.AppEnv)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:
		return err
	case <-ctx.Done():
		slog.Info("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}
