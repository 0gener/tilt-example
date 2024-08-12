package awsmessaging

import (
	"github.com/0gener/go-service/components"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// WithAWSEndpoint overrides the default AWS endpoint. This is particularly useful for local testing.
func WithAWSEndpoint(endpoint string) components.Option {
	return func(component components.Component) error {
		messagingComponent, err := components.AsComponent[*Component](component)
		if err != nil {
			return err
		}

		messagingComponent.snsConfigOpts = append(messagingComponent.snsConfigOpts, func(options *sns.Options) {
			options.BaseEndpoint = aws.String(endpoint)
			options.Credentials = credentials.NewStaticCredentialsProvider("dummy", "dummy", "")
		})

		messagingComponent.sqsConfigOpts = append(messagingComponent.sqsConfigOpts, func(options *sqs.Options) {
			options.BaseEndpoint = aws.String(endpoint)
			options.Credentials = credentials.NewStaticCredentialsProvider("dummy", "dummy", "")
		})

		return nil
	}
}
