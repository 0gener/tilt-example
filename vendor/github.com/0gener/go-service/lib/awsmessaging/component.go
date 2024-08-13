package awsmessaging

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/0gener/go-service/components"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	snstypes "github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	sqstypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/google/uuid"
)

const (
	ComponentName = "awsmessaging"
)

type awsNotification struct {
	Message           string                           `json:"Message"`
	MessageAttributes map[string]NotificationAttribute `json:"MessageAttributes"`
}

type NotificationAttribute struct {
	Type  string `json:"Type"`
	Value string `json:"Value"`
}

type Component struct {
	components.BaseComponent

	snsClient *sns.Client
	sqsClient *sqs.Client

	snsConfigOpts []func(*sns.Options)
	sqsConfigOpts []func(*sqs.Options)
}

func New() *Component {
	return &Component{
		BaseComponent: *components.NewBaseComponent(ComponentName),
	}
}

func (component *Component) Configure(ctx context.Context) error {
	sdkConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return err
	}

	component.snsClient = sns.NewFromConfig(sdkConfig, func(options *sns.Options) {
		for _, opt := range component.snsConfigOpts {
			opt(options)
		}
	})

	component.sqsClient = sqs.NewFromConfig(sdkConfig, func(options *sqs.Options) {
		for _, opt := range component.sqsConfigOpts {
			opt(options)
		}
	})

	component.NotifyStatus(components.CONFIGURED)
	return nil
}

func (component *Component) Publish(ctx context.Context, topicArn string, message *Message) error {
	_, err := component.snsClient.Publish(ctx, &sns.PublishInput{
		Message:           aws.String(string(message.Data)),
		MessageAttributes: extractAwsMessageAttributes(*message),
		TopicArn:          aws.String(topicArn),
	})
	return err
}

func (component *Component) CreateTopic(ctx context.Context, topic string) (string, error) {
	res, err := component.snsClient.CreateTopic(ctx, &sns.CreateTopicInput{
		Name: aws.String(topic),
	})
	if err != nil {
		return "", err
	}

	return *res.TopicArn, nil
}

func (component *Component) CreateQueueForTopic(ctx context.Context, topicArn string, queueName string) (string, error) {
	res, err := component.sqsClient.CreateQueue(ctx, &sqs.CreateQueueInput{
		QueueName: aws.String(queueName),
	})
	if err != nil {
		return "", err
	}

	attr, err := component.sqsClient.GetQueueAttributes(ctx, &sqs.GetQueueAttributesInput{
		QueueUrl: res.QueueUrl,
		AttributeNames: []sqstypes.QueueAttributeName{
			"QueueArn",
		},
	})
	if err != nil {
		return "", err
	}

	queueArn := attr.Attributes["QueueArn"]

	_, err = component.snsClient.Subscribe(ctx, &sns.SubscribeInput{
		Protocol:              aws.String("sqs"),
		TopicArn:              aws.String(topicArn),
		Endpoint:              aws.String(queueArn),
		ReturnSubscriptionArn: false,
	})
	if err != nil {
		return "", err
	}

	return *res.QueueUrl, nil
}

func (component *Component) Subscribe(queueUrl string, handler HandlerFunc, opts ...SubOpt) *Subscription {
	return newSubscription(component.sqsClient, component.Logger(), queueUrl, handler, opts...)
}

func extractMessage(msg sqstypes.Message) (*Message, error) {
	var notification awsNotification
	err := json.Unmarshal([]byte(*msg.Body), &notification)
	if err != nil {
		return nil, err
	}

	awsMessage := notification.Message
	awsAttributes := notification.MessageAttributes

	return &Message{
		ID:         uuid.New(),
		Data:       []byte(awsMessage),
		Attributes: extractMessageAttributes(awsAttributes),
		Err:        nil,
	}, nil
}

func extractMessageAttributes(awsAttrs map[string]NotificationAttribute) MessageAttributes {
	attr := make(MessageAttributes)
	for key, value := range awsAttrs {
		attr[key] = value.Value
	}
	return attr
}

func extractAwsMessageAttributes(msg Message) map[string]snstypes.MessageAttributeValue {
	messageAttributes := make(map[string]snstypes.MessageAttributeValue)

	for key, val := range msg.Attributes {
		messageAttributes[key] = snstypes.MessageAttributeValue{
			DataType:    aws.String("String"),
			StringValue: aws.String(val),
		}
	}

	return messageAttributes
}

func validateQueueExists(ctx context.Context, sqsClient *sqs.Client, queueUrl string) error {
	_, err := sqsClient.GetQueueAttributes(ctx, &sqs.GetQueueAttributesInput{
		QueueUrl: aws.String(queueUrl),
	})
	if err != nil {
		return fmt.Errorf("failed to get queue attributes: %w", err)
	}

	return nil
}
