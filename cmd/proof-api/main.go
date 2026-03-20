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

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/statictoad/vtc-platform/internal/proof"
	"github.com/statictoad/vtc-platform/internal/proof/upstream"
	"github.com/statictoad/vtc-platform/internal/shared/cache"
	"github.com/statictoad/vtc-platform/internal/shared/db"
	"github.com/statictoad/vtc-platform/pkg/events"
)

func main() {
	// -------------------------------------------------------------------------
	// Logger
	// -------------------------------------------------------------------------
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// -------------------------------------------------------------------------
	// Config
	// -------------------------------------------------------------------------
	port := getEnv("PORT", "3004")

	operator := proof.OperatorConfig{
		Name:       getEnv("OPERATOR_NAME", ""),
		Siret:      getEnv("OPERATOR_SIRET", ""),
		EvtcNumber: getEnv("OPERATOR_EVTC_NUMBER", ""),
	}

	// -------------------------------------------------------------------------
	// Database
	// -------------------------------------------------------------------------
	pool, err := db.Connect(context.Background())
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	slog.Info("database connected")

	// -------------------------------------------------------------------------
	// Valkey
	// -------------------------------------------------------------------------
	valkeyClient, err := cache.NewClient()
	if err != nil {
		slog.Error("failed to connect to valkey", "error", err)
		os.Exit(1)
	}
	defer valkeyClient.Close()

	slog.Info("valkey connected")

	// -------------------------------------------------------------------------
	// Dependencies
	// -------------------------------------------------------------------------
	repo := proof.NewRepository(pool)

	clientSvc := upstream.NewClientServiceClient(getEnv("CLIENT_SERVICE_URL", "http://localhost:3002"))
	fleetSvc := upstream.NewFleetServiceClient(getEnv("FLEET_SERVICE_URL", "http://localhost:3003"))

	svc, err := proof.NewService(repo, operator, clientSvc, fleetSvc)
	if err != nil {
		slog.Error("failed to initialise proof service", "error", err)
		os.Exit(1)
	}

	handler := proof.NewHandler(svc)

	// -------------------------------------------------------------------------
	// Event consumer — booking.confirmed
	// -------------------------------------------------------------------------
	consumer := cache.NewConsumer(valkeyClient, "proof-service", "proof-service-1")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := consumer.EnsureGroup(ctx, events.StreamBookingConfirmed); err != nil {
		slog.Error("failed to ensure consumer group", "error", err)
		os.Exit(1)
	}

	go func() {
		slog.Info("proof consumer started", "stream", events.StreamBookingConfirmed)
		if err := consumer.Consume(ctx, events.StreamBookingConfirmed, svc.HandleBookingConfirmed); err != nil {
			slog.Error("consumer error", "error", err)
		}
	}()

	// -------------------------------------------------------------------------
	// Router
	// -------------------------------------------------------------------------
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler.RegisterRoutes(r)

	// -------------------------------------------------------------------------
	// Server
	// -------------------------------------------------------------------------
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("proof-api starting", "port", port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	// -------------------------------------------------------------------------
	// Graceful shutdown
	// -------------------------------------------------------------------------
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down...")

	// Cancel consumer context first — stops the Valkey consumer loop.
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("forced shutdown", "error", err)
		os.Exit(1)
	}

	slog.Info("proof-api stopped")
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
