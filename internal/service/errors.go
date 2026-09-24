package service

import "errors"

var (
	ErrEventNotFound     = errors.New("event not found")
	ErrSeatAlreadyExists = errors.New("seat already exists")
	ErrSeatAlreadyBooked = errors.New("seat already booked")
)
