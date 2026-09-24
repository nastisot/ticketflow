package service

import (
	"context"
	"ticketflow/internal/domain"
	"ticketflow/internal/models"
	"time"
)

type EventRepository interface {
	GetAll(ctx context.Context) ([]models.Event, error)
	GetByID(ctx context.Context, id int64) (*models.Event, error)
	Create(ctx context.Context, name string, address string, date time.Time) (*models.Event, error)
}

type SeatRepository interface {
	Create(ctx context.Context, eventID int64, number string, priceCents int64) (*models.Seat, error)
	GetByEventID(ctx context.Context, eventID int64) ([]models.SeatWithAvailability, error)
}

type BookingRepository interface {
	Create(ctx context.Context, booking domain.Booking) (*domain.Booking, error)
	GetByID(ctx context.Context, id int64) (*domain.Booking, error)
	Confirm(ctx context.Context, id int64) (*domain.Booking, error)
	Expire(ctx context.Context, id int64, from domain.BookingStatus) (*domain.Booking, error)
	Cancel(ctx context.Context, id int64, from domain.BookingStatus) (*domain.Booking, error)
	GetExpiredPendingIDs(ctx context.Context, limit int) ([]int64, error)
}
