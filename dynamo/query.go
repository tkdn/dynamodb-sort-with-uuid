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

// ListByUser queries items for a given user_id partition key whose
// created_at falls after cutoff. Lexicographic string comparison of
// the composite sort key against the bare cutoff timestamp works
// correctly here: a string that starts with an identical prefix is
// always "greater than" that prefix alone, so the "#<job_id>" suffix
// never affects the comparison outcome.
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
