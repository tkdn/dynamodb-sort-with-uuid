package dynamo

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

const (
	localEndpoint = "http://localhost:8000"
	localRegion   = "us-east-1"
)

// NewClient builds a DynamoDB client hardcoded to talk to a local
// DynamoDB emulator (amazon/dynamodb-local). Not for production use:
// credentials and endpoint are intentionally static, since this
// project never talks to real AWS.
func NewClient(ctx context.Context) (*dynamodb.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(localRegion),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider("dummy", "dummy", ""),
		),
	)
	if err != nil {
		return nil, err
	}

	client := dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
		o.BaseEndpoint = new(localEndpoint)
	})

	return client, nil
}
