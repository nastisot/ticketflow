package service

import (
	"context"
	"ticketflow/internal/models"
	"time"
)

type EventService struct {
	repo EventRepository
}

func NewEventService(repo EventRepository) *EventService {
	return &EventService{repo: repo}
}

func (s *EventService) GetAll(ctx context.Context) ([]models.Event, error) {
	return s.repo.GetAll(ctx)
}

func (s *EventService) GetByID(ctx context.Context, id int64) (*models.Event, error) {
	event, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, ErrEventNotFound
	}
	return event, nil
}

func (s *EventService) Create(ctx context.Context, name string, address string, date time.Time) (*models.Event, error) {
	return s.repo.Create(ctx, name, address, date)
}
