package domain

import "time"

type OutboxEvent struct {
	ID          int64
	EventType   string
	AggregateID int64
	Payload     []byte
	CreatedAt   time.Time
	PublishedAt *time.Time
}
