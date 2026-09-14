package service

import "errors"

var (
	ErrEventNotFound     = errors.New("event not found")
	ErrSeatAlreadyExists = errors.New("seat already exists")
	ErrSeatNotFound      = errors.New("seat not found")
	ErrUserNotFound      = errors.New("user not found")
	ErrSeatAlreadyBooked = errors.New("seat already booked")
	ErrBookingNotFound   = errors.New("booking not found")
)
