package awsmessaging

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	sqstypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"time"
)

const (
	startupTimeout = 10 * time.Second

	defaultWorkers                    = 10
	defaultMaxNumberOfMessages        = 10
	defaultVisibilityTimeoutInSeconds = 2
	defaultWaitTimeInSeconds          = 10
)

type subscriptionConfig struct {
	workers                    int32
	maxNumberOfMessages        int32
	visibilityTimeoutInSeconds int32
	waitTimeInSeconds          int32
}

type Subscription struct {
	logger    *zap.Logger
	handler   HandlerFunc
	queueUrl  string
	cfg       *subscriptionConfig
	sqs       *sqs.Client
	semaphore chan struct{}
	ctx       context.Context
	cancel    context.CancelFunc
}

func WithWorkers(workers int32) SubOpt {
	return func(cfg *subscriptionConfig) {
		cfg.workers = workers
	}
}

func WithMaxNumberOfMessages(maxNumberOfMessages int32) SubOpt {
	return func(cfg *subscriptionConfig) {
		cfg.maxNumberOfMessages = maxNumberOfMessages
	}
}

func WithVisibilityTimeoutInSeconds(visibilityTimeoutInSeconds int32) SubOpt {
	return func(cfg *subscriptionConfig) {
		cfg.visibilityTimeoutInSeconds = visibilityTimeoutInSeconds
	}
}

func WithWaitTimeInSeconds(waitTimeInSeconds int32) SubOpt {
	return func(cfg *subscriptionConfig) {
		cfg.waitTimeInSeconds = waitTimeInSeconds
	}
}

type SubOpt func(cfg *subscriptionConfig)

type HandlerFunc func(message []*Message)

func newSubscription(sqs *sqs.Client, logger *zap.Logger, queueUrl string, handler HandlerFunc, opts ...SubOpt) *Subscription {
	cfg := loadSubscriptionConfig(opts...)
	ctx, cancel := context.WithCancel(context.Background())

	return &Subscription{
		logger:    logger,
		sqs:       sqs,
		handler:   handler,
		queueUrl:  queueUrl,
		cfg:       cfg,
		semaphore: make(chan struct{}, cfg.workers),
		ctx:       ctx,
		cancel:    cancel,
	}
}

func (s *Subscription) Start() error {
	startupCtx, startUpCancel := context.WithTimeout(s.ctx, startupTimeout)
	defer startUpCancel()

	if err := validateQueueExists(startupCtx, s.sqs, s.queueUrl); err != nil {
		s.cancel()
		return err
	}

	go func() {
		for {
			select {
			case <-s.ctx.Done():
				return
			default:
				out, err := s.sqs.ReceiveMessage(s.ctx, s.getReceiveMessagesRequest(s.queueUrl))
				if err != nil {
					s.logger.Error("failed to receive messages", zap.Error(err))
					continue
				}

				if len(out.Messages) > 0 {
					s.semaphore <- struct{}{}
					go s.handleBatch(out.Messages)
				}
			}
		}
	}()

	return nil
}

func (s *Subscription) Stop() {
	s.cancel()
}

func (s *Subscription) handleBatch(awsMessages []sqstypes.Message) {
	defer func() {
		<-s.semaphore
	}()

	var messages []*Message
	deleteMessagesMap := make(map[uuid.UUID]sqstypes.DeleteMessageBatchRequestEntry)
	for _, awsMessage := range awsMessages {
		message, err := extractMessage(awsMessage)
		if err != nil {
			s.logger.Warn("failed to extract message", zap.Error(err))
			continue
		}

		messages = append(messages, message)
		deleteMessagesMap[message.ID] = sqstypes.DeleteMessageBatchRequestEntry{
			Id:            aws.String(message.ID.String()),
			ReceiptHandle: awsMessage.ReceiptHandle,
		}
	}

	s.handler(messages)

	var successMessagesDeleteEntries []sqstypes.DeleteMessageBatchRequestEntry
	for _, message := range messages {
		if message.Err != nil {
			s.logger.Warn("failed to process message", zap.Error(message.Err))
			continue
		}

		successMessagesDeleteEntries = append(successMessagesDeleteEntries, deleteMessagesMap[message.ID])
	}

	if len(successMessagesDeleteEntries) > 0 {
		_, err := s.sqs.DeleteMessageBatch(s.ctx, &sqs.DeleteMessageBatchInput{
			QueueUrl: aws.String(s.queueUrl),
			Entries:  successMessagesDeleteEntries,
		})
		if err != nil {
			s.logger.Error("failed to delete messages", zap.Error(err))
		}
	}
}

func (s *Subscription) getReceiveMessagesRequest(queueUrl string) *sqs.ReceiveMessageInput {
	return &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(queueUrl),
		MaxNumberOfMessages: s.cfg.maxNumberOfMessages,
		VisibilityTimeout:   s.cfg.visibilityTimeoutInSeconds,
		WaitTimeSeconds:     s.cfg.waitTimeInSeconds,
	}
}

func loadSubscriptionConfig(opts ...SubOpt) *subscriptionConfig {
	cfg := &subscriptionConfig{
		workers:                    defaultWorkers,
		maxNumberOfMessages:        defaultMaxNumberOfMessages,
		visibilityTimeoutInSeconds: defaultVisibilityTimeoutInSeconds,
		waitTimeInSeconds:          defaultWaitTimeInSeconds,
	}

	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}
