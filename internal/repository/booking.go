package repository

import (
	"context"
	"errors"
	"ticketflow/internal/models"
	"ticketflow/internal/service"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BookingRepository struct {
	db *pgxpool.Pool
}

func NewBookingRepository(db *pgxpool.Pool) *BookingRepository {
	return &BookingRepository{db: db}
}

func (r *BookingRepository) Create(ctx context.Context, seatID int64, userID int64) (*models.Booking, error) {
	var booking models.Booking
	err := r.db.QueryRow(ctx, `
		INSERT INTO bookings (seat_id, user_id)
    	VALUES ($1, $2)
        RETURNING id, seat_id, user_id, created_at;`, seatID, userID).Scan(&booking.ID, &booking.SeatID, &booking.UserID, &booking.CreatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				if pgErr.ConstraintName == "bookings_seat_id_key" {
					return nil, service.ErrSeatAlreadyBooked
				}
			case "23503":
				if pgErr.ConstraintName == "bookings_seat_id_fkey" {
					return nil, service.ErrSeatNotFound
				}
				if pgErr.ConstraintName == "bookings_user_id_fkey" {
					return nil, service.ErrUserNotFound
				}
			}
		}
		return nil, err
	}
	return &booking, nil
}

func (r *BookingRepository) GetByID(ctx context.Context, id int64) (*models.Booking, error) {
	var booking models.Booking
	err := r.db.QueryRow(ctx, `
		SELECT id, seat_id, user_id, created_at
		FROM bookings
		WHERE id = $1`, id).Scan(&booking.ID, &booking.SeatID, &booking.UserID, &booking.CreatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}
	return &booking, nil
}

func (r *BookingRepository) Delete(ctx context.Context, id int64) (bool, error) {
	result, err := r.db.Exec(ctx, `DELETE FROM bookings WHERE id = $1`, id)
	if err != nil {
		return false, err
	}
	return result.RowsAffected() > 0, nil
}
