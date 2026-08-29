package dynamo

import (
	"encoding/json/v2"
	"fmt"
	"strings"
	"time"
	"uuid"
)

// sortKeyTimeLayoutは、辞書順ソートが時系列順と一致するよう、固定桁数・
// ゼロ埋めのナノ秒表記を生成する。time.RFC3339Nanoは".999999999"パターンで
// 末尾のゼロを削り可変長になるため、意図的に使わない（辞書順ソートと
// ソートキーへの範囲クエリの両方が壊れる）。
const sortKeyTimeLayout = "2006-01-02T15:04:05.000000000Z"

// LookbackWindowは、ListByUserがItemをどこまで遡って取得するかを定める。
// サーバー側の固定カットオフとして適用される（クライアントからは変更不可）。
const LookbackWindow = 24 * time.Hour

// Itemは、DynamoDBの"items"テーブルの行を表す。
// SortKeyは"<created_at>#<job_id>"という生のソートキー値を保持し、
// JSON出力からは除外される。CreatedAt()がAPIレスポンス用に時刻部分を
// 再構成する。
type Item struct {
	UserID  string `dynamodbav:"user_id"`
	SortKey string `dynamodbav:"created_at"`
	JobID   string `dynamodbav:"job_id"`
}

// CreatedAtは、複合ソートキーから時刻部分を取り出す。
func (i Item) CreatedAt() string {
	createdAt, _, _ := strings.Cut(i.SortKey, "#")
	return createdAt
}

// MarshalJSONは、API向けの形（user_id, created_at, job_id）を出力する。
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

// NewItemは、"現在時刻"のItemを組み立てる。標準uuidパッケージで新規job_idを
// 生成し、created_atを固定桁数のナノ秒精度でフォーマットする。
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
