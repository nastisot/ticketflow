package worker

import (
	"context"
	"log/slog"
	"ticketflow/internal/domain"
	"time"
)

type EventPublisher interface {
	Publish(ctx context.Context, event domain.OutboxEvent) error
}

type OutboxRepository interface {
	ClaimUnpublished(ctx context.Context, limit int) ([]domain.OutboxEvent, error)
	MarkPublished(ctx context.Context, id int64) error
}

type OutboxWorker struct {
	repo           OutboxRepository
	publisher      EventPublisher
	logger         *slog.Logger
	interval       time.Duration
	publishTimeout time.Duration
	limit          int
}

func NewOutboxWorker(repo OutboxRepository, publisher EventPublisher, logger *slog.Logger, interval time.Duration, publishTimeout time.Duration, limit int) *OutboxWorker {
	return &OutboxWorker{
		repo:           repo,
		publisher:      publisher,
		logger:         logger,
		interval:       interval,
		publishTimeout: publishTimeout,
		limit:          limit,
	}
}

func (w *OutboxWorker) processBatch(ctx context.Context) error {
	events, err := w.repo.ClaimUnpublished(ctx, w.limit)
	if err != nil {
		return err
	}
	for _, event := range events {
		publishCtx, cancel := context.WithTimeout(ctx, w.publishTimeout)
		err = w.publisher.Publish(publishCtx, event)

		cancel()

		if err != nil {
			return err
		}
		err = w.repo.MarkPublished(ctx, event.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (w *OutboxWorker) Run(ctx context.Context) {
	w.logger.Info(
		"outbox worker started",
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
			if err := w.processBatch(ctx); err != nil {
				if ctx.Err() != nil {
					return
				}
				w.logger.Error(
					"failed to process outbox events",
					"error", err,
				)
			}
		}
	}
}
