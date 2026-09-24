package domain

import "errors"

var (
	ErrInvalidBookingTransition = errors.New("invalid booking status transition")
	ErrBookingNotFound          = errors.New("booking not found")
	ErrBookingExpired           = errors.New("booking expired")
	ErrSeatAlreadyBooked        = errors.New("seat already booked")
	ErrSeatNotFound             = errors.New("seat not found")
	ErrUserNotFound             = errors.New("user not found")
	ErrBookingStateConflict     = errors.New("booking state changed")
	ErrBookingNotExpired        = errors.New("booking has not expired yet")
)
