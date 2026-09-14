//go:build integration

package repository_test

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"
	"ticketflow/internal/repository"
	"ticketflow/internal/service"

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

func TestBookingRepository_ConcurrentCreate(t *testing.T) {
	db := openTestDB(t)

	_, err := db.Exec(context.Background(), `
		TRUNCATE bookings, seats, users, events
		RESTART IDENTITY CASCADE
	`)
	if err != nil {
		t.Fatal(err)
	}

	var eventID int64
	err = db.QueryRow(context.Background(), `
		INSERT INTO events (name, address, date)
		VALUES ('Test Event', 'Test Address', NOW())
		RETURNING id
		`).Scan(&eventID)
	if err != nil {
		t.Fatal(err)
	}

	var seatID int64
	err = db.QueryRow(context.Background(), `
 		INSERT INTO seats (event_id, number, price_cents)
 		VALUES ($1, 'A1', 10000)
 		RETURNING id
 		`, eventID).Scan(&seatID)
	if err != nil {
		t.Fatal(err)
	}

	var user1ID int64
	var user2ID int64

	err = db.QueryRow(context.Background(), `
        INSERT INTO users (name)
        VALUES ('User 1')
        RETURNING id
        `).Scan(&user1ID)
	if err != nil {
		t.Fatal(err)
	}

	err = db.QueryRow(context.Background(), `
        INSERT INTO users (name)
        VALUES ('User 2')
        RETURNING id
        `).Scan(&user2ID)
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup

	results := make(chan error, 2)
	start := make(chan struct{})

	repo := repository.NewBookingRepository(db)

	wg.Add(2)

	go func() {
		defer wg.Done()
		<-start
		_, err := repo.Create(context.Background(), seatID, user1ID)
		results <- err
	}()

	go func() {
		defer wg.Done()
		<-start
		_, err := repo.Create(context.Background(), seatID, user2ID)
		results <- err
	}()
	close(start)

	wg.Wait()
	close(results)

	successCount := 0
	conflictCount := 0

	for err := range results {
		if err == nil {
			successCount++
		}
		if errors.Is(err, service.ErrSeatAlreadyBooked) {
			conflictCount++
		}
	}
	if successCount != 1 || conflictCount != 1 {
		t.Fatalf("got %d results, want %d results", successCount, conflictCount)
	}

	var count int64
	err = db.QueryRow(context.Background(), `SELECT COUNT(*) FROM bookings WHERE seat_id = $1`, seatID).Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("got row %d, want 1", count)
	}
}
