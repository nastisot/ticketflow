package models

import "time"

type Event struct {
	ID      int64     `json:"id"`
	Name    string    `json:"name"`
	Address string    `json:"address"`
	Date    time.Time `json:"date"`
}

type Seat struct {
	ID         int64  `json:"id"`
	EventID    int64  `json:"event_id"`
	Number     string `json:"number"`
	PriceCents int64  `json:"price_cents"`
}

type SeatWithAvailability struct {
	ID         int64  `json:"id"`
	EventID    int64  `json:"event_id"`
	Number     string `json:"number"`
	PriceCents int64  `json:"price_cents"`
	Available  bool   `json:"available"`
}

type User struct {
	ID   int64
	Name string
}

type Booking struct {
	ID        int64
	SeatID    int64
	UserID    int64
	CreatedAt time.Time
}
