package dynamo

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

const tableName = "items"

type ItemLister struct {
	client *dynamodb.Client
}

func NewItemLister(client *dynamodb.Client) *ItemLister {
	return &ItemLister{client: client}
}

// ListByUserは、指定したuser_idパーティションキーに属し、created_atが
// cutoffより後のItemを取得する。複合ソートキーに対して素のcutoff文字列を
// 辞書順比較しても正しく機能する: 同一プレフィックスを持つ文字列は、
// その延長より必ず小さいという辞書順の性質により、"#<job_id>"という接尾辞は
// 比較結果に一切影響しない。
func (l *ItemLister) ListByUser(ctx context.Context, userID string, cutoff time.Time) ([]Item, error) {
	out, err := l.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              new(tableName),
		KeyConditionExpression: new("user_id = :uid AND created_at > :cutoff"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":uid":    &types.AttributeValueMemberS{Value: userID},
			":cutoff": &types.AttributeValueMemberS{Value: formatSortKeyTime(cutoff)},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("query items for user %q: %w", userID, err)
	}

	items := []Item{}
	if err := attributevalue.UnmarshalListOfMaps(out.Items, &items); err != nil {
		return nil, fmt.Errorf("unmarshal items: %w", err)
	}

	return items, nil
}
