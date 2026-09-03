package events

import (
	"context"
	providers "ecommerce/internal/providers"
	"encoding/json"
	"fmt"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-aws/sqs"
	"github.com/ThreeDotsLabs/watermill/message"

	appconfig "ecommerce/internal/config"

	_ "github.com/aws/smithy-go/endpoints"
)

type EventPublisher struct {
	publisher message.Publisher
	queueName string
}

func (ep *EventPublisher) Publish(eventType string, payload interface{}, metadata map[string]string) error {

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	msg := message.NewMessage(watermill.NewUUID(), data)

	// Add metadata
	msg.Metadata.Set("event_type", eventType)
	for k, v := range metadata {
		msg.Metadata.Set(k, v)
	}

	return ep.publisher.Publish(ep.queueName, msg)
}

func (ep *EventPublisher) Close() error {
	return ep.publisher.Close()
}

func NewEventPublisher(ctx context.Context, cfg *appconfig.AWSConfig) (*EventPublisher, error) {
	logger := watermill.NewStdLogger(false, false)

	// Create AWS config
	awsConfig, err := providers.CreateAWSConfig(ctx, cfg.S3Endpoint, cfg.Region)
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS config: %w", err)
	}

	// Create Watermill SQS publisher config
	publisherConfig := sqs.PublisherConfig{
		AWSConfig: awsConfig,
		Marshaler: nil,
	}

	// Create the publisher with custom config
	publisher, err := sqs.NewPublisher(publisherConfig, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create publisher: %w", err)
	}

	return &EventPublisher{
		publisher: publisher,
		queueName: cfg.EventQueueName,
	}, nil
}
