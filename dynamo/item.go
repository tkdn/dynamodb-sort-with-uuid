package dynamo

import (
	"encoding/json/v2"
	"fmt"
	"strings"
	"time"
	"uuid"
)

// sortKeyTimeLayout produces a fixed-width, zero-padded nanosecond
// fraction so lexicographic string sort matches chronological order.
// Deliberately NOT time.RFC3339Nano, which uses ".999999999" and
// trims trailing zeros (variable width, breaks lexicographic sort
// and breaks range queries against the sort key).
const sortKeyTimeLayout = "2006-01-02T15:04:05.000000000Z"

// LookbackWindow bounds how far back ListByUser will look for items,
// applied server-side as a fixed cutoff (not client-configurable).
const LookbackWindow = 24 * time.Hour

// Item mirrors the DynamoDB "items" table row.
// SortKey holds the raw "<created_at>#<job_id>" sort-key value and is
// excluded from JSON output; CreatedAt() reconstructs the timestamp
// portion for API responses.
type Item struct {
	UserID  string `dynamodbav:"user_id"`
	SortKey string `dynamodbav:"created_at"`
	JobID   string `dynamodbav:"job_id"`
}

// CreatedAt extracts the timestamp portion of the composite sort key.
func (i Item) CreatedAt() string {
	createdAt, _, _ := strings.Cut(i.SortKey, "#")
	return createdAt
}

// MarshalJSON renders the API-facing shape: user_id, created_at, job_id.
func (i Item) MarshalJSON() ([]byte, error) {
	type response struct {
		UserID    string `json:"user_id"`
		CreatedAt string `json:"created_at"`
		JobID     string `json:"job_id"`
	}
	return json.Marshal(response{
		UserID:    i.UserID,
		CreatedAt: i.CreatedAt(),
		JobID:     i.JobID,
	})
}

// NewItem builds an Item for "now", generating a fresh job_id via the
// stdlib uuid package and formatting created_at with fixed-width
// nanosecond precision.
func NewItem(userID string) Item {
	now := time.Now().UTC()
	createdAt := formatSortKeyTime(now)
	jobID := uuid.New().String()

	return Item{
		UserID:  userID,
		SortKey: composeSortKey(createdAt, jobID),
		JobID:   jobID,
	}
}

func formatSortKeyTime(t time.Time) string {
	return t.UTC().Format(sortKeyTimeLayout)
}

func composeSortKey(createdAt, jobID string) string {
	return fmt.Sprintf("%s#%s", createdAt, jobID)
}
