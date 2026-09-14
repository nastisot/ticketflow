package repository

import (
	"context"
	"errors"
	"ticketflow/internal/models"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EventRepository struct {
	db *pgxpool.Pool
}

func NewEventRepository(db *pgxpool.Pool) *EventRepository {
	return &EventRepository{db: db}
}

func (r *EventRepository) GetAll(ctx context.Context) ([]models.Event, error) {
	rows, err := r.db.Query(ctx, "SELECT id, name, address, date FROM events ORDER BY id;")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	events := make([]models.Event, 0)
	for rows.Next() {
		var event models.Event
		err = rows.Scan(&event.ID, &event.Name, &event.Address, &event.Date)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return events, nil
}

func (r *EventRepository) GetByID(ctx context.Context, id int64) (*models.Event, error) {
	var event models.Event
	err := r.db.QueryRow(ctx, "SELECT id, name, address, date FROM events WHERE id=$1;", id).Scan(&event.ID, &event.Name, &event.Address, &event.Date)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *EventRepository) Create(ctx context.Context, name string, address string, date time.Time) (*models.Event, error) {
	var event models.Event
	err := r.db.QueryRow(ctx,
		`INSERT INTO events (name, address, date)
			 VALUES ($1, $2, $3)
			 RETURNING id, name, address, date;`,
		name, address, date).Scan(&event.ID, &event.Name, &event.Address, &event.Date)
	if err != nil {
		return nil, err
	}
	return &event, nil
}
