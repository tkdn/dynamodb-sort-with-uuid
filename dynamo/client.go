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

// NewClientは、ローカルのDynamoDBエミュレータ（amazon/dynamodb-local）に接続する
// DynamoDBクライアントを構築する。本番利用は想定していない。認証情報とエンドポイントは
// 意図的に固定値であり、このプロジェクトは実AWSには接続しない。
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
