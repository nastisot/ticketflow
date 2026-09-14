package service

import (
	"context"
	"ticketflow/internal/models"
)

type SeatService struct {
	seatRepo  SeatRepository
	eventRepo EventRepository
}

func NewSeatService(seatRepo SeatRepository, eventRepo EventRepository) *SeatService {
	return &SeatService{
		seatRepo:  seatRepo,
		eventRepo: eventRepo,
	}
}

func (s *SeatService) Create(ctx context.Context, eventID int64, number string, priceCents int64) (*models.Seat, error) {
	return s.seatRepo.Create(ctx, eventID, number, priceCents)
}

func (s *SeatService) GetByEventID(ctx context.Context, eventID int64) ([]models.SeatWithAvailability, error) {
	event, err := s.eventRepo.GetByID(ctx, eventID)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, ErrEventNotFound
	}

	return s.seatRepo.GetByEventID(ctx, eventID)
}
