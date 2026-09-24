package domain

import (
	"time"
)

type BookingStatus string

const (
	BookingStatusPending   BookingStatus = "pending"
	BookingStatusConfirmed BookingStatus = "confirmed"
	BookingStatusCancelled BookingStatus = "cancelled"
	BookingStatusExpired   BookingStatus = "expired"
)

type Booking struct {
	ID        int64         `json:"id"`
	SeatID    int64         `json:"seat_id"`
	UserID    int64         `json:"user_id"`
	Status    BookingStatus `json:"status"`
	ExpiresAt *time.Time    `json:"expires_at"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}

var bookingTransitions = map[BookingStatus]map[BookingStatus]struct{}{
	BookingStatusPending: {
		BookingStatusConfirmed: {},
		BookingStatusCancelled: {},
		BookingStatusExpired:   {},
	},

	BookingStatusConfirmed: {
		BookingStatusCancelled: {},
	},
}

func ValidateBookingTransition(from BookingStatus, to BookingStatus) error {
	allowedTransitions, ok := bookingTransitions[from]
	if !ok {
		return ErrInvalidBookingTransition
	}
	if _, ok := allowedTransitions[to]; !ok {
		return ErrInvalidBookingTransition
	}
	return nil
}
