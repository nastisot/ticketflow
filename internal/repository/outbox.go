package repository

import (
	"context"
	"ticketflow/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type OutboxRepository struct {
	db pgxpool.Pool
}

func NewOutboxRepository(db pgxpool.Pool) *OutboxRepository {
	return &OutboxRepository{db: db}
}
func (r *OutboxRepository) GetUnpublished(ctx context.Context, limit int) ([]domain.OutboxEvent, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, event_type, aggregate_id, payload, created_at, published_at
        FROM outbox_events
        WHERE published_at IS NULL
        ORDER BY created_at ASC, id ASC
        LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var events []domain.OutboxEvent
	for rows.Next() {
		var event domain.OutboxEvent
		if err := rows.Scan(
			&event.ID,
			&event.EventType,
			&event.AggregateID,
			&event.Payload,
			&event.CreatedAt,
			&event.PublishedAt,
		); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return events, nil
}

func (r *OutboxRepository) MarkPublished(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx, `
		UPDATE outbox_events
		SET published_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND published_at IS NULL`, id)
	if err != nil {
		return err
	}
	return nil
}
