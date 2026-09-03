package events

import (
	"context"
	"errors"
	"fmt"

	watermillsqs "github.com/ThreeDotsLabs/watermill-aws/sqs"
	"github.com/aws/aws-sdk-go-v2/aws"
	awssqs "github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/rs/zerolog"

	appconfig "ecommerce/internal/config"
	"ecommerce/internal/providers"
)

// NewSQSEventPublisher wires an EventPublisher on top of Amazon SQS. A custom
// endpoint (LocalStack) also forces the SQS client to that address, exactly
// like the S3 provider does for object storage.
func NewSQSEventPublisher(
	ctx context.Context,
	cfg *appconfig.Config,
	logger zerolog.Logger,
) (*EventPublisher, error) {
	if cfg.AWS.EventQueueName == "" {
		return nil, errors.New("events: AWS_EVENT_QUEUE_NAME is not set")
	}

	awsCfg, err := providers.CreateAWSConfig(ctx, cfg.AWS.SQSEndpoint, cfg.AWS.Region)
	if err != nil {
		return nil, fmt.Errorf("failed to load aws config: %w", err)
	}

	publisher, err := watermillsqs.NewPublisher(watermillsqs.PublisherConfig{
		AWSConfig: awsCfg,
		OptFns: []func(*awssqs.Options){
			func(options *awssqs.Options) {
				if cfg.AWS.SQSEndpoint != "" {
					options.BaseEndpoint = aws.String(cfg.AWS.SQSEndpoint)
				}
			},
		},
	}, NewWatermillLogger(logger))
	if err != nil {
		return nil, fmt.Errorf("failed to create sqs publisher: %w", err)
	}

	return NewEventPublisher(publisher, cfg.AWS.EventQueueName)
}
