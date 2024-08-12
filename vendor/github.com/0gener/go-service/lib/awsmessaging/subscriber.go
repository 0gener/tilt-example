package awsmessaging

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
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

type HandlerFunc func(message *Message) error

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

				for _, msg := range out.Messages {
					s.semaphore <- struct{}{}
					go s.handleMessage(msg)
				}
			}
		}
	}()

	return nil
}

func (s *Subscription) Stop() {
	s.cancel()
}

func (s *Subscription) handleMessage(msg awstypes.Message) {
	defer func() {
		<-s.semaphore
	}()

	s.logger.Debug("received message", zap.Any("message", msg))

	message, err := extractMessage(msg)
	if err != nil {
		s.logger.Error("failed to extract message", zap.Error(err))
		return
	}

	err = s.handler(message)
	if err == nil {
		_, err = s.sqs.DeleteMessage(context.TODO(), &sqs.DeleteMessageInput{
			QueueUrl:      aws.String(s.queueUrl),
			ReceiptHandle: msg.ReceiptHandle,
		})
		if err != nil {
			s.logger.Error("failed to delete message", zap.Error(err))
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
