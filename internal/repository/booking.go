package repository

import (
	"context"
	"errors"
	"ticketflow/internal/domain"

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

func (r *BookingRepository) Create(ctx context.Context, booking domain.Booking) (*domain.Booking, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var seatID int64

	err = tx.QueryRow(ctx, `
		SELECT id
        FROM seats
        WHERE id = $1
        FOR UPDATE`,
		booking.SeatID).Scan(&seatID)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrSeatNotFound
	}
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(ctx, `
		UPDATE bookings 
        SET status = 'expired', updated_at = CURRENT_TIMESTAMP
        WHERE seat_id = $1 AND status = 'pending' AND expires_at <= CURRENT_TIMESTAMP`,
		booking.SeatID)

	if err != nil {
		return nil, err
	}

	var bookingNew domain.Booking

	err = tx.QueryRow(ctx, `
		INSERT INTO bookings (seat_id, user_id, status, expires_at)
    	VALUES ($1, $2, $3, $4)
        RETURNING id, seat_id, user_id, status, expires_at, created_at, updated_at;`,
		booking.SeatID,
		booking.UserID,
		booking.Status,
		booking.ExpiresAt,
	).Scan(
		&bookingNew.ID,
		&bookingNew.SeatID,
		&bookingNew.UserID,
		&bookingNew.Status,
		&bookingNew.ExpiresAt,
		&bookingNew.CreatedAt,
		&bookingNew.UpdatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				if pgErr.ConstraintName == "bookings_active_seat_unique" {
					return nil, domain.ErrSeatAlreadyBooked
				}
			case "23503":
				if pgErr.ConstraintName == "bookings_user_id_fkey" {
					return nil, domain.ErrUserNotFound
				}
			}
		}
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &bookingNew, nil
}

func (r *BookingRepository) GetByID(ctx context.Context, id int64) (*domain.Booking, error) {
	var booking domain.Booking
	err := r.db.QueryRow(ctx, `
		SELECT id, seat_id, user_id, status, expires_at, created_at, updated_at
		FROM bookings
		WHERE id = $1;`, id).Scan(
		&booking.ID,
		&booking.SeatID,
		&booking.UserID,
		&booking.Status,
		&booking.ExpiresAt,
		&booking.CreatedAt,
		&booking.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}
	return &booking, nil
}

func (r *BookingRepository) Confirm(ctx context.Context, id int64) (*domain.Booking, error) {
	var booking domain.Booking
	err := r.db.QueryRow(ctx, `
		UPDATE bookings
        SET status = 'confirmed', updated_at = CURRENT_TIMESTAMP WHERE id = $1 AND status = 'pending' AND expires_at > CURRENT_TIMESTAMP
        RETURNING id, seat_id, user_id, status, expires_at, created_at, updated_at;`,
		id,
	).Scan(
		&booking.ID,
		&booking.SeatID,
		&booking.UserID,
		&booking.Status,
		&booking.ExpiresAt,
		&booking.CreatedAt,
		&booking.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &booking, nil
}

func (r *BookingRepository) Expire(ctx context.Context, id int64, from domain.BookingStatus) (*domain.Booking, error) {
	var booking domain.Booking
	err := r.db.QueryRow(ctx, `
		UPDATE bookings
        SET status = 'expired', updated_at = CURRENT_TIMESTAMP 
        WHERE id = $1 AND status = $2 AND expires_at <= CURRENT_TIMESTAMP
        RETURNING id, seat_id, user_id, status, expires_at, created_at, updated_at;`,
		id,
		from,
	).Scan(
		&booking.ID,
		&booking.SeatID,
		&booking.UserID,
		&booking.Status,
		&booking.ExpiresAt,
		&booking.CreatedAt,
		&booking.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &booking, nil
}

func (r *BookingRepository) Cancel(ctx context.Context, id int64, from domain.BookingStatus) (*domain.Booking, error) {
	var booking domain.Booking
	err := r.db.QueryRow(ctx, `
		UPDATE bookings
        SET status = 'cancelled', updated_at = CURRENT_TIMESTAMP
        WHERE id = $1 AND status = $2
        RETURNING id, seat_id, user_id, status, expires_at, created_at, updated_at;`,
		id,
		from,
	).Scan(
		&booking.ID,
		&booking.SeatID,
		&booking.UserID,
		&booking.Status,
		&booking.ExpiresAt,
		&booking.CreatedAt,
		&booking.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &booking, nil
}

func (r *BookingRepository) GetExpiredPendingIDs(ctx context.Context, limit int) ([]int64, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id
		FROM bookings
		WHERE status = 'pending' AND expires_at <= CURRENT_TIMESTAMP
		ORDER BY expires_at
		LIMIT $1;`, limit)

	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ids := make([]int64, 0)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return ids, nil
}
