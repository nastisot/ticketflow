package worker

import (
	"context"
	"log/slog"
	"ticketflow/internal/service"
	"time"
)

type BookingExpirationWorker struct {
	service  *service.BookingService
	logger   *slog.Logger
	interval time.Duration
	limit    int
}

func NewBookingExpirationWorker(service *service.BookingService, logger *slog.Logger, interval time.Duration, limit int) *BookingExpirationWorker {
	return &BookingExpirationWorker{
		service:  service,
		logger:   logger,
		interval: interval,
		limit:    limit,
	}
}

func (w *BookingExpirationWorker) Run(ctx context.Context) {
	w.logger.Info(
		"booking expiration worker started",
		"interval", w.interval,
		"limit", w.limit,
	)

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := w.service.ExpirePending(ctx, w.limit); err != nil {
				if ctx.Err() != nil {
					return
				}
				w.logger.Error(
					"failed to expire pending bookings",
					"error", err,
				)
			}
		}
	}
}
