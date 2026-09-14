package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"ticketflow/internal/config"
	"ticketflow/internal/handler"
	"ticketflow/internal/middleware"
	"ticketflow/internal/repository"
	"ticketflow/internal/service"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		logger.Error(
			"failed to load config",
			"error", err,
		)
		os.Exit(1)
	}

	db, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error(
			"failed to connect database pool",
			"error", err,
		)
		os.Exit(1)
	}

	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		logger.Error(
			"database is unavailable",
			"error", err)
		os.Exit(1)
	}

	eventRepo := repository.NewEventRepository(db)
	eventService := service.NewEventService(eventRepo)
	eventHandler := handler.NewEventHandler(eventService, logger)

	seatRepo := repository.NewSeatRepository(db)
	seatService := service.NewSeatService(seatRepo, eventRepo)
	seatHandler := handler.NewSeatHandler(seatService, logger)

	bookingRepo := repository.NewBookingRepository(db)
	bookingService := service.NewBookingService(bookingRepo)
	bookingHandler := handler.NewBookingHandler(bookingService, logger)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /events", eventHandler.GetAll)
	mux.HandleFunc("GET /events/{id}", eventHandler.GetByID)
	mux.HandleFunc("POST /events", eventHandler.Create)
	mux.HandleFunc("POST /events/{id}/seats", seatHandler.Create)
	mux.HandleFunc("GET /events/{id}/seats", seatHandler.GetByEventID)
	mux.HandleFunc("POST /seats/{id}/bookings", bookingHandler.Create)
	mux.HandleFunc("GET /bookings/{id}", bookingHandler.GetByID)
	mux.HandleFunc("DELETE /bookings/{id}", bookingHandler.Delete)

	httpHandler := middleware.RequestID(middleware.Logging(logger, mux))

	serverErrorLogger := slog.NewLogLogger(logger.Handler(), slog.LevelError)

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: httpHandler,

		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,

		ErrorLog: serverErrorLogger,
	}

	logger.Info(
		"starting server",
		"port", cfg.Port,
	)

	go func() {
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error(
				"server failed",
				"error", err,
			)
		}
	}()

	shutdownSignalCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-shutdownSignalCtx.Done()

	logger.Info("shutting down server")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error(
			"failed to shutdown server",
			"error", err,
		)
		return
	}
	logger.Info("server stopped")
}
