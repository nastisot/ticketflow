//go:build integration

package repository_test

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"
	"ticketflow/internal/domain"
	"ticketflow/internal/repository"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func openTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Fatalf("TEST_DATABASE_URL is not set")
	}

	ctx := context.Background()

	db, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Ping(ctx); err != nil {
		db.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.Close()
	})
	return db

}

func createTestEvent(t *testing.T, db *pgxpool.Pool) int64 {
	t.Helper()

	var id int64

	err := db.QueryRow(context.Background(), `
		INSERT INTO events (name, address, date)
		VALUES ('Test Event', 'Test Address', NOW())
		RETURNING id
		`).Scan(&id)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func createTestSeat(t *testing.T, db *pgxpool.Pool, eventID int64, number string) int64 {
	t.Helper()

	var id int64

	err := db.QueryRow(context.Background(), `
 		INSERT INTO seats (event_id, number, price_cents)
 		VALUES ($1, $2, 10000)
 		RETURNING id
 		`, eventID, number).Scan(&id)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func createTestUser(t *testing.T, db *pgxpool.Pool, name string) int64 {
	t.Helper()

	var id int64

	err := db.QueryRow(context.Background(), `
        INSERT INTO users (name)
        VALUES ($1)
        RETURNING id
        `, name).Scan(&id)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func cleanTestDB(t *testing.T, db *pgxpool.Pool) {
	t.Helper()

	_, err := db.Exec(context.Background(), `
		TRUNCATE bookings, seats, users, events
		RESTART IDENTITY CASCADE
	`)
	if err != nil {
		t.Fatal(err)
	}
}

func TestBookingRepository_ConcurrentCreate(t *testing.T) {
	db := openTestDB(t)

	cleanTestDB(t, db)

	eventID := createTestEvent(t, db)
	seatID := createTestSeat(t, db, eventID, "A1")

	user1ID := createTestUser(t, db, "User 1")
	user2ID := createTestUser(t, db, "User 2")

	repo := repository.NewBookingRepository(db)

	var wg sync.WaitGroup

	results := make(chan error, 2)
	start := make(chan struct{})

	wg.Add(2)

	go func(user1ID int64) {
		defer wg.Done()
		<-start

		expiresAt := time.Now().Add(10 * time.Minute)

		_, err := repo.Create(context.Background(), domain.Booking{
			SeatID:    seatID,
			UserID:    user1ID,
			Status:    domain.BookingStatusPending,
			ExpiresAt: &expiresAt,
		})

		results <- err
	}(user1ID)

	go func(user2ID int64) {
		defer wg.Done()
		<-start

		expiresAt := time.Now().Add(10 * time.Minute)
		_, err := repo.Create(context.Background(), domain.Booking{
			SeatID:    seatID,
			UserID:    user2ID,
			Status:    domain.BookingStatusPending,
			ExpiresAt: &expiresAt,
		})

		results <- err
	}(user2ID)

	close(start)

	wg.Wait()
	close(results)

	successCount := 0
	conflictCount := 0

	for err := range results {
		switch {
		case err == nil:
			successCount++
		case errors.Is(err, domain.ErrSeatAlreadyBooked):
			conflictCount++
		default:
			t.Fatalf("unexpected error: %v", err)
		}
	}

	if successCount != 1 || conflictCount != 1 {
		t.Fatalf("expected 1 success and 1 conflict, got %d success and %d conflicts", successCount, conflictCount)
	}

	var count int64
	err := db.QueryRow(context.Background(), `SELECT COUNT(*) FROM bookings WHERE seat_id = $1 AND status IN ('pending', 'confirmed')`, seatID).Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("got row %d, want 1", count)
	}
}

func TestBookingStalePending(t *testing.T) {
	db := openTestDB(t)

	cleanTestDB(t, db)

	eventID := createTestEvent(t, db)
	seatID := createTestSeat(t, db, eventID, "A1")

	user1ID := createTestUser(t, db, "User 1")

	repo := repository.NewBookingRepository(db)

	_, err := db.Exec(context.Background(), `
		INSERT INTO bookings (seat_id, user_id, status, expires_at)
		VALUES ($1, $2, 'pending', CURRENT_TIMESTAMP - INTERVAL '1 minute')`, seatID, user1ID)
	if err != nil {
		t.Fatal(err)
	}

	expiresAt := time.Now().Add(10 * time.Minute)

	booking := domain.Booking{
		SeatID:    seatID,
		UserID:    user1ID,
		Status:    domain.BookingStatusPending,
		ExpiresAt: &expiresAt,
	}
	created, err := repo.Create(context.Background(), booking)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if created.Status != domain.BookingStatusPending {
		t.Fatalf("expected pending, got %s", created.Status)
	}

	rows, err := db.Query(context.Background(), `SELECT status FROM bookings WHERE seat_id = $1 ORDER BY id`, seatID)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	var statuses []domain.BookingStatus

	for rows.Next() {
		var status domain.BookingStatus

		if err := rows.Scan(&status); err != nil {
			t.Fatal(err)
		}
		statuses = append(statuses, status)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}

	if len(statuses) != 2 || statuses[0] != domain.BookingStatusExpired || statuses[1] != domain.BookingStatusPending {
		t.Fatalf("expected [expired pending], got %v", statuses)
	}
}
