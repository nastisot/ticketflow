//go:build integration

package repository_test

import (
	"context"
	"sync"
	"testing"
	"ticketflow/internal/domain"
	"ticketflow/internal/repository"
	"time"
)

func TestOutboxRepository_ClaimUnpublished(t *testing.T) {
	db := openTestDB(t)
	cleanTestDB(t, db)

	eventID := createTestEvent(t, db)
	seatID := createTestSeat(t, db, eventID, "A1")
	userID := createTestUser(t, db, "User 1")

	bookingRepo := repository.NewBookingRepository(db)

	expiresAt := time.Now().Add(10 * time.Minute)

	booking, err := bookingRepo.Create(context.Background(), domain.Booking{
		SeatID:    seatID,
		UserID:    userID,
		Status:    domain.BookingStatusPending,
		ExpiresAt: &expiresAt,
	})
	if err != nil {
		t.Fatalf("create booking: %v", err)
	}

	for i := 0; i < 4; i++ {
		_, err := db.Exec(context.Background(), `
			INSERT INTO outbox_events (event_type, aggregate_id, payload)
			VALUES ('booking.confirmed', $1, '{}'::jsonb)`, booking.ID)
		if err != nil {
			t.Fatal(err)
		}
	}

	repo := repository.NewOutboxRepository(db)

	start := make(chan struct{})

	results := make(chan []domain.OutboxEvent, 2)
	errs := make(chan error, 2)

	var wg sync.WaitGroup
	wg.Add(2)

	claim := func() {
		defer wg.Done()

		<-start

		events, err := repo.ClaimUnpublished(context.Background(), 2)
		if err != nil {
			errs <- err
			return
		}
		results <- events
	}

	go claim()
	go claim()

	close(start)

	wg.Wait()
	close(results)
	close(errs)
	for err := range errs {
		t.Fatalf("ClaimUnpublished: %v", err)
	}

	claimedIDs := make(map[int64]bool)
	total := 0

	for events := range results {
		for _, event := range events {
			total++
			if claimedIDs[event.ID] {
				t.Fatalf("event %d was claimed more than once", event.ID)
			}
			claimedIDs[event.ID] = true
		}
	}
	if total != 4 {
		t.Fatalf("expected 4 claimed events, got %d", total)
	}

	var processingCount int64

	err = db.QueryRow(context.Background(), `
		SELECT COUNT(*) FROM outbox_events
		WHERE processing_at IS NOT NULL AND published_at IS NULL`).Scan(&processingCount)
	if err != nil {
		t.Fatal(err)
	}
	if processingCount != 4 {
		t.Fatalf("expected 4 processing events, got %d", processingCount)
	}
}

func TestOutboxRepository_DoesNotReclaimProcessingEvent(t *testing.T) {
	db := openTestDB(t)
	cleanTestDB(t, db)

	eventID := createTestEvent(t, db)
	seatID := createTestSeat(t, db, eventID, "A1")
	userID := createTestUser(t, db, "User 1")

	bookingRepo := repository.NewBookingRepository(db)

	expiresAt := time.Now().Add(10 * time.Minute)
	booking, err := bookingRepo.Create(context.Background(), domain.Booking{
		SeatID:    seatID,
		UserID:    userID,
		Status:    domain.BookingStatusPending,
		ExpiresAt: &expiresAt,
	})
	if err != nil {
		t.Fatalf("create booking: %v", err)
	}

	_, err = db.Exec(context.Background(), `
		INSERT INTO outbox_events (event_type, aggregate_id, payload)
		VALUES ('booking.confirmed', $1, '{}'::jsonb)`, booking.ID)
	if err != nil {
		t.Fatal(err)
	}

	repo := repository.NewOutboxRepository(db)

	firstClaim, err := repo.ClaimUnpublished(context.Background(), 1)
	if err != nil {
		t.Fatalf("first claim: %v", err)
	}

	if len(firstClaim) != 1 {
		t.Fatalf("expected 1 claim, got %d", len(firstClaim))
	}

	secondClaim, err := repo.ClaimUnpublished(context.Background(), 1)
	if err != nil {
		t.Fatalf("second claim: %v", err)
	}

	if len(secondClaim) != 0 {
		t.Fatalf("expected 0 events on second claim, got %d", len(secondClaim))
	}
}

func TestOutboxRepository_ReclaimsStaleProcessingEvent(t *testing.T) {
	db := openTestDB(t)
	cleanTestDB(t, db)

	eventID := createTestEvent(t, db)
	seatID := createTestSeat(t, db, eventID, "A1")
	userID := createTestUser(t, db, "User 1")

	bookingRepo := repository.NewBookingRepository(db)
	expiresAt := time.Now().Add(10 * time.Minute)

	booking, err := bookingRepo.Create(context.Background(), domain.Booking{
		SeatID:    seatID,
		UserID:    userID,
		Status:    domain.BookingStatusPending,
		ExpiresAt: &expiresAt,
	})
	if err != nil {
		t.Fatalf("create booking: %v", err)
	}

	_, err = db.Exec(context.Background(), `
		INSERT INTO outbox_events (event_type, aggregate_id, payload)
    	VALUES ('booking.confirmed', $1, '{}'::jsonb)`, booking.ID)
	if err != nil {
		t.Fatal(err)
	}

	repo := repository.NewOutboxRepository(db)

	firstClaim, err := repo.ClaimUnpublished(context.Background(), 1)
	if err != nil {
		t.Fatalf("first claim: %v", err)
	}
	if len(firstClaim) != 1 {
		t.Fatalf("expected 1 event, got %d", len(firstClaim))
	}

	outboxID := firstClaim[0].ID

	_, err = db.Exec(context.Background(), `
		UPDATE outbox_events
		SET processing_at = CURRENT_TIMESTAMP - INTERVAL '2 minutes'
		WHERE id = $1`, outboxID)

	secondClaim, err := repo.ClaimUnpublished(context.Background(), 1)
	if err != nil {
		t.Fatalf("second claim: %v", err)
	}
	if len(secondClaim) != 1 {
		t.Fatalf("expected stale event to be reclaimed, got %d events", len(secondClaim))
	}
	if secondClaim[0].ID != outboxID {
		t.Fatalf("expected event %d, got %d", outboxID, secondClaim[0].ID)
	}
}

func TestOutboxRepository_MarkPublishedRemovesEventFromClaim(t *testing.T) {
	db := openTestDB(t)
	cleanTestDB(t, db)

	eventID := createTestEvent(t, db)
	seatID := createTestSeat(t, db, eventID, "A1")
	userID := createTestUser(t, db, "User 1")

	bookingRepo := repository.NewBookingRepository(db)
	expiresAt := time.Now().Add(10 * time.Minute)

	booking, err := bookingRepo.Create(context.Background(), domain.Booking{
		SeatID:    seatID,
		UserID:    userID,
		Status:    domain.BookingStatusPending,
		ExpiresAt: &expiresAt,
	})
	if err != nil {
		t.Fatalf("create booking: %v", err)
	}

	_, err = db.Exec(context.Background(), `
		INSERT INTO outbox_events (event_type, aggregate_id, payload)
    	VALUES ('booking.confirmed', $1, '{}'::jsonb)`, booking.ID)

	repo := repository.NewOutboxRepository(db)

	claimed, err := repo.ClaimUnpublished(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(claimed) != 1 {
		t.Fatalf("expected 1 claim, got %d", len(claimed))
	}

	event1ID := claimed[0].ID

	err = repo.MarkPublished(context.Background(), event1ID)
	if err != nil {
		t.Fatalf("MarkPublished: %v", err)
	}

	claimedAgain, err := repo.ClaimUnpublished(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}

	if len(claimedAgain) != 0 {
		t.Fatalf("expected published event not to be claimed again, got %d events", len(claimedAgain))
	}

	var publishedAt *time.Time

	err = db.QueryRow(context.Background(), `
		SELECT published_at FROM outbox_events
		WHERE id = $1`, event1ID).Scan(&publishedAt)
	if err != nil {
		t.Fatal(err)
	}
	if publishedAt == nil {
		t.Fatal("expected published_at to be set")
	}
}
