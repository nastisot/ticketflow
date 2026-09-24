package service

import (
	"context"
	"errors"
	"ticketflow/internal/domain"
	"time"
)

type BookingService struct {
	repo       BookingRepository
	bookingTTL time.Duration
}

func NewBookingService(repo BookingRepository, bookingTTL time.Duration) *BookingService {
	return &BookingService{
		repo:       repo,
		bookingTTL: bookingTTL,
	}
}

func (s *BookingService) Create(ctx context.Context, seatID int64, userID int64) (*domain.Booking, error) {
	expiresAt := time.Now().Add(s.bookingTTL)

	booking := domain.Booking{
		SeatID:    seatID,
		UserID:    userID,
		Status:    domain.BookingStatusPending,
		ExpiresAt: &expiresAt,
	}
	return s.repo.Create(ctx, booking)
}

func (s *BookingService) GetByID(ctx context.Context, id int64) (*domain.Booking, error) {
	booking, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if booking == nil {
		return nil, domain.ErrBookingNotFound
	}
	return booking, nil
}

func (s *BookingService) Confirm(ctx context.Context, id int64) (*domain.Booking, error) {
	booking, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if booking == nil {
		return nil, domain.ErrBookingNotFound
	}

	if err := domain.ValidateBookingTransition(booking.Status, domain.BookingStatusConfirmed); err != nil {
		return nil, err
	}

	if booking.ExpiresAt == nil {
		return nil, errors.New("pending booking has no expiration time")
	}
	now := time.Now()

	if !now.Before(*booking.ExpiresAt) {
		return nil, domain.ErrBookingExpired
	}

	updatedBooking, err := s.repo.Confirm(ctx, id)
	if err != nil {
		return nil, err
	}
	if updatedBooking == nil {
		return nil, domain.ErrBookingStateConflict
	}
	return updatedBooking, nil

}

func (s *BookingService) Expire(ctx context.Context, id int64) (*domain.Booking, error) {
	booking, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if booking == nil {
		return nil, domain.ErrBookingNotFound
	}
	if err := domain.ValidateBookingTransition(booking.Status, domain.BookingStatusExpired); err != nil {
		return nil, err
	}
	if booking.ExpiresAt == nil {
		return nil, errors.New("pending booking has no expiration time")
	}
	now := time.Now()
	if now.Before(*booking.ExpiresAt) {
		return nil, domain.ErrBookingNotExpired
	}
	expiredBooking, err := s.repo.Expire(ctx, id, booking.Status)
	if err != nil {
		return nil, err
	}
	if expiredBooking == nil {
		return nil, domain.ErrBookingStateConflict
	}
	return expiredBooking, nil
}

func (s *BookingService) Cancel(ctx context.Context, id int64) (*domain.Booking, error) {
	booking, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if booking == nil {
		return nil, domain.ErrBookingNotFound
	}
	if err := domain.ValidateBookingTransition(booking.Status, domain.BookingStatusCancelled); err != nil {
		return nil, err
	}
	cancelBooking, err := s.repo.Cancel(ctx, id, booking.Status)
	if err != nil {
		return nil, err
	}
	if cancelBooking == nil {
		return nil, domain.ErrBookingStateConflict
	}
	return cancelBooking, nil
}

func (s *BookingService) ExpirePending(ctx context.Context, limit int) error {
	ids, err := s.repo.GetExpiredPendingIDs(ctx, limit)
	if err != nil {
		return err
	}
	for _, id := range ids {
		_, err := s.Expire(ctx, id)
		if err == nil {
			continue
		}
		if errors.Is(err, domain.ErrBookingNotFound) || errors.Is(err, domain.ErrBookingStateConflict) || errors.Is(err, domain.ErrInvalidBookingTransition) || errors.Is(err, domain.ErrBookingNotExpired) {
			continue
		}
		return err
	}
	return nil
}
