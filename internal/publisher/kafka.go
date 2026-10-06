package publisher

import (
	"context"
	"strconv"
	"ticketflow/internal/domain"

	"github.com/twmb/franz-go/pkg/kgo"
)

type KafkaPublisher struct {
	client *kgo.Client
	topic  string
}

func NewKafkaPublisher(brokers []string, topic string) (*KafkaPublisher, error) {
	client, err := kgo.NewClient(kgo.SeedBrokers(brokers...))
	if err != nil {
		return nil, err
	}

	return &KafkaPublisher{
		client: client,
		topic:  topic,
	}, nil
}

func (p *KafkaPublisher) Publish(ctx context.Context, event domain.OutboxEvent) error {
	record := &kgo.Record{
		Topic: p.topic,
		Key:   []byte(strconv.FormatInt(event.AggregateID, 10)),
		Value: event.Payload,
	}
	return p.client.ProduceSync(ctx, record).FirstErr()
}

func (p *KafkaPublisher) Close() {
	p.client.Close()
}
