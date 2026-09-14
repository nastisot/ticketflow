package service

import (
	"context"
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
	Create(ctx context.Context, seatID int64, userID int64) (*models.Booking, error)
	GetByID(ctx context.Context, id int64) (*models.Booking, error)
	Delete(ctx context.Context, id int64) (bool, error)
}
