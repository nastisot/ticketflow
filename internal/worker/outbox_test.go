package worker

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"ticketflow/internal/domain"
	"time"
)

type fakePublisher struct {
	published []domain.OutboxEvent
	err       error
}

func (p *fakePublisher) Publish(ctx context.Context, event domain.OutboxEvent) error {
	if p.err != nil {
		return p.err
	}
	p.published = append(p.published, event)
	return nil
}

type fakeOutboxRepository struct {
	events    []domain.OutboxEvent
	markedIDs []int64
}

func (r *fakeOutboxRepository) ClaimUnpublished(ctx context.Context, limit int) ([]domain.OutboxEvent, error) {
	return r.events, nil
}

func (r *fakeOutboxRepository) MarkPublished(ctx context.Context, id int64) error {
	r.markedIDs = append(r.markedIDs, id)
	return nil
}

func TestOutboxWorker_ProcessBatch(t *testing.T) {
	event := domain.OutboxEvent{
		ID:          1,
		EventType:   domain.BookingConfirmedEventType,
		AggregateID: 42,
		Payload:     []byte(`{"booking_id":42}`),
	}
	repo := &fakeOutboxRepository{
		events: []domain.OutboxEvent{event},
	}

	publisher := &fakePublisher{}

	worker := NewOutboxWorker(repo, publisher, slog.Default(), time.Second, 100)

	err := worker.processBatch(context.Background())
	if err != nil {
		t.Fatalf("processBatch: %v", err)
	}

	if len(publisher.published) != 1 {
		t.Fatalf("len(publisher.published) = %d, want 1", len(publisher.published))
	}

	if publisher.published[0].ID != event.ID {
		t.Fatalf("publisher.published[0].ID = %d, want %d", publisher.published[0].ID, event.ID)
	}

	if len(repo.markedIDs) != 1 {
		t.Fatalf("len(repo.markedIDs) = %d, want 1", len(repo.markedIDs))
	}

	if repo.markedIDs[0] != event.ID {
		t.Fatalf("repo.markedIDs[0] = %d, want %d", repo.markedIDs[0], event.ID)
	}
}

func TestOutboxWorker_DoesNotMarkPublishedWhenPublishFails(t *testing.T) {
	event := domain.OutboxEvent{
		ID:          1,
		EventType:   domain.BookingConfirmedEventType,
		AggregateID: 42,
		Payload:     []byte(`{"booking_id":42}`),
	}
	repo := &fakeOutboxRepository{
		events: []domain.OutboxEvent{event},
	}
	publisher := &fakePublisher{
		err: errors.New("kafka unavailable"),
	}

	worker := NewOutboxWorker(repo, publisher, slog.Default(), time.Second, 100)

	err := worker.processBatch(context.Background())

	if err == nil {
		t.Fatal("expected error")
	}

	if len(repo.markedIDs) != 0 {
		t.Fatalf("expected no marked events, got %d", len(repo.markedIDs))
	}
}
