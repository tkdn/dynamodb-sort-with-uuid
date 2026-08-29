package handler

import (
	"context"
	"encoding/json/v2"
	"net/http"
	"time"

	"github.com/tkdn/dynamodb-sort-uuid/dynamo"
)

// ItemListerは、ハンドラが必要とする最小限のインターフェース。AWS SDKに
// 依存せずに、ハンドラ層のテストでフェイクに差し替えられるようにする。
type ItemLister interface {
	ListByUser(ctx context.Context, userID string, cutoff time.Time) ([]dynamo.Item, error)
}

// dynamoCallTimeoutは、DynamoDB呼び出し1回がブロックしてよい時間の上限。
// これがないと、到達不能なDynamoDBエンドポイント（例: 停止したローカル
// コンテナ）に対して、想定よりはるかに長くブロックしうる。環境によっては
// TCP接続の試行がすぐには失敗しないため、SDK呼び出しが妥当なリクエスト
// タイムアウトを大きく超えてブロックすることがある。
const dynamoCallTimeout = 5 * time.Second

type ItemsHandler struct {
	lister ItemLister
}

func NewItemsHandler(lister ItemLister) *ItemsHandler {
	return &ItemsHandler{lister: lister}
}

// Listは、GET /users/{user_id}/itemsを処理する。保持窓はサーバー側の
// 固定カットオフであり、クライアントは異なる範囲を要求できない。
func (h *ItemsHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("user_id")
	// 空のuser_idは通常のルーティング経路では到達しない: ServeMuxの
	// パスクリーニングにより"/users//items"は"/users/items"
	// （3セグメントパターンとの不一致）にリダイレクトされ、このハンドラには
	// 到達する前に処理が終わる。将来の入力経路の変化（例えば異なる形の
	// ルート）に備えた保険として残している。
	if userID == "" {
		writeError(w, http.StatusBadRequest, "user_id is required")
		return
	}

	cutoff := time.Now().UTC().Add(-dynamo.LookbackWindow)

	ctx, cancel := context.WithTimeout(r.Context(), dynamoCallTimeout)
	defer cancel()

	items, err := h.lister.ListByUser(ctx, userID, cutoff)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list items")
		return
	}

	writeJSON(w, http.StatusOK, items)
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.MarshalWrite(w, body)
}

type errorResponse struct {
	Error string `json:"error"`
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorResponse{Error: msg})
}
