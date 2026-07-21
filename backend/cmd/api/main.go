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
	paygw "github.com/alamjad/marketplace/internal/platform/payment"
	"github.com/alamjad/marketplace/internal/platform/presence"
	"github.com/alamjad/marketplace/internal/platform/storage"
	"github.com/alamjad/marketplace/internal/repository/postgres"
	authuc "github.com/alamjad/marketplace/internal/usecase/auth"
	"github.com/alamjad/marketplace/internal/usecase/catalog"
	contactuc "github.com/alamjad/marketplace/internal/usecase/contact"
	listinguc "github.com/alamjad/marketplace/internal/usecase/listing"
	paymentuc "github.com/alamjad/marketplace/internal/usecase/payment"
	settingsuc "github.com/alamjad/marketplace/internal/usecase/settings"
	useruc "github.com/alamjad/marketplace/internal/usecase/user"
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
	payProvider, err := newPaymentProvider(cfg)
	if err != nil {
		return err
	}
	slog.Info("payment provider", "provider", payProvider.Name())
	repo := postgres.New(pool)

	// Usecases.
	authService := authuc.NewService(repo, tokens)
	catalogService := catalog.NewService(repo)
	paymentService := paymentuc.NewService(repo, payProvider, cfg.FrontendBaseURL)
	listingService := listinguc.NewService(repo)
	userService := useruc.NewService(repo)
	contactService := contactuc.NewService(repo)
	settingsService := settingsuc.NewService(repo)

	// Live presence: a session is "online" for 60s after its last heartbeat.
	presenceTracker := presence.NewTracker(60 * time.Second)

	// Background sweeper: retire listings whose month is up.
	go runExpirySweeper(ctx, listingService)

	// HTTP layer.
	router := nethttp.NewRouter(nethttp.RouterDeps{
		AllowedOrigins: cfg.CORSAllowedOrigins,
		UploadDir:      cfg.UploadDir,
		Auth:           middleware.NewAuthenticator(tokens, userService),
		Health:         handler.NewHealthHandler(pool),
		AuthH:          handler.NewAuthHandler(authService),
		Catalog:        handler.NewCatalogHandler(catalogService),
		Listing:        handler.NewListingHandler(listingService, store, paymentService),
		User:           handler.NewUserHandler(userService),
		Payment:        handler.NewPaymentHandler(paymentService),
		Contact:        handler.NewContactHandler(contactService),
		Settings:       handler.NewSettingsHandler(settingsService),
		Presence:       handler.NewPresenceHandler(presenceTracker),
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

// newPaymentProvider selects the payment gateway from config. "thawani" enables
// the real Thawani Pay checkout (requires keys); anything else uses the dev stub.
func newPaymentProvider(cfg *config.Config) (paygw.Provider, error) {
	switch cfg.PaymentProvider {
	case "thawani":
		return paygw.NewThawani(cfg.ThawaniBaseURL, cfg.ThawaniSecretKey, cfg.ThawaniPublishableKey)
	default:
		return paygw.NewStub(cfg.FrontendBaseURL), nil
	}
}

// runExpirySweeper flips month-old active listings to 'expired', once at
// startup then hourly, until the context is canceled.
func runExpirySweeper(ctx context.Context, svc *listinguc.Service) {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	for {
		if n, err := svc.ExpireDue(ctx); err != nil {
			if ctx.Err() == nil {
				slog.Error("expiry sweep failed", "error", err)
			}
		} else if n > 0 {
			slog.Info("expired listings", "count", n)
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
