package service

import (
	"context"
	"ticketflow/internal/models"
)

type BookingService struct {
	repo BookingRepository
}

func NewBookingService(repo BookingRepository) *BookingService {
	return &BookingService{repo: repo}
}

func (s *BookingService) Create(ctx context.Context, seatID int64, userID int64) (*models.Booking, error) {
	return s.repo.Create(ctx, seatID, userID)
}

func (s *BookingService) GetByID(ctx context.Context, id int64) (*models.Booking, error) {
	booking, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if booking == nil {
		return nil, ErrBookingNotFound
	}
	return booking, nil
}

func (s *BookingService) Delete(ctx context.Context, id int64) error {
	deleted, err := s.repo.Delete(ctx, id)
	if err != nil {
		return err
	}
	if !deleted {
		return ErrBookingNotFound
	}

	return nil
}
