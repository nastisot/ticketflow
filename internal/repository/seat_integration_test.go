//go:build integration

package repository_test

import (
	"context"
	"testing"
	"ticketflow/internal/repository"
)

func TestSeatRepository_AvailabilityWithStalePending(t *testing.T) {
	db := openTestDB(t)
	cleanTestDB(t, db)

	eventID := createTestEvent(t, db)
	seatID := createTestSeat(t, db, eventID, "A1")
	userID := createTestUser(t, db, "User 1")

	_, err := db.Exec(context.Background(), `
		INSERT INTO bookings (seat_id, user_id, status, expires_at)
		VALUES ($1, $2, 'pending', CURRENT_TIMESTAMP + INTERVAL '10 minutes')`, seatID, userID)
	if err != nil {
		t.Fatal(err)
	}

	repo := repository.NewSeatRepository(db)

	seats, err := repo.GetByEventID(context.Background(), eventID)
	if err != nil {
		t.Fatal(err)
	}
	if len(seats) != 1 {
		t.Fatalf("got %d seats, expected 1", len(seats))
	}
	if seats[0].Available {
		t.Fatal("expected seat to be unavailable")
	}

	_, err = db.Exec(context.Background(), `
		UPDATE bookings
        SET expires_at = CURRENT_TIMESTAMP - INTERVAL '1 minute'
        WHERE seat_id = $1`, seatID)

	if err != nil {
		t.Fatal(err)
	}
	seats, err = repo.GetByEventID(context.Background(), eventID)
	if err != nil {
		t.Fatal(err)
	}
	if len(seats) != 1 {
		t.Fatalf("got %d seats, expected 1", len(seats))
	}
	if !seats[0].Available {
		t.Fatal("expected seat to be available")
	}
}
