package repository

import (
	"context"
	"errors"
	"ticketflow/internal/models"
	"ticketflow/internal/service"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SeatRepository struct {
	db *pgxpool.Pool
}

func NewSeatRepository(db *pgxpool.Pool) *SeatRepository {
	return &SeatRepository{db: db}
}

func (r *SeatRepository) Create(ctx context.Context, eventID int64, number string, priceCents int64) (*models.Seat, error) {
	var seat models.Seat
	err := r.db.QueryRow(ctx,
		`INSERT INTO seats (event_id, number, price_cents)
             VALUES ($1, $2, $3)
             RETURNING id, event_id, number, price_cents;`, eventID, number, priceCents).Scan(&seat.ID, &seat.EventID, &seat.Number, &seat.PriceCents)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				return nil, service.ErrSeatAlreadyExists
			case "23503":
				return nil, service.ErrEventNotFound
			}
		}
		return nil, err
	}
	return &seat, nil
}

func (r *SeatRepository) GetByEventID(ctx context.Context, eventID int64) ([]models.SeatWithAvailability, error) {
	rows, err := r.db.Query(ctx,
		`SELECT s.id, s.event_id, s.number, s.price_cents, b.id IS NULL AS available
	         FROM seats s LEFT JOIN bookings b ON b.seat_id = s.id AND b.status IN ('pending', 'confirmed')
	         WHERE s.event_id = $1
	         ORDER BY s.id;`, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	seats := make([]models.SeatWithAvailability, 0)
	for rows.Next() {
		var seat models.SeatWithAvailability
		err = rows.Scan(&seat.ID, &seat.EventID, &seat.Number, &seat.PriceCents, &seat.Available)
		if err != nil {
			return nil, err
		}
		seats = append(seats, seat)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return seats, nil
}
