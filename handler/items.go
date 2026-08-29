package handler

import (
	"context"
	"encoding/json/v2"
	"net/http"
	"time"

	"github.com/tkdn/dynamodb-sort-uuid/dynamo"
)

// ItemLister is the minimal interface the handler needs, allowing a
// fake to be substituted in handler-level tests without pulling in
// the AWS SDK.
type ItemLister interface {
	ListByUser(ctx context.Context, userID string, cutoff time.Time) ([]dynamo.Item, error)
}

// dynamoCallTimeout bounds how long a single DynamoDB call may block.
// Without this, an unreachable DynamoDB endpoint (e.g. a stopped local
// container) can hang far longer than expected: TCP connection attempts
// on some networks do not fail fast, so the SDK call can block well
// past what a caller would consider a reasonable request timeout.
const dynamoCallTimeout = 5 * time.Second

type ItemsHandler struct {
	lister ItemLister
}

func NewItemsHandler(lister ItemLister) *ItemsHandler {
	return &ItemsHandler{lister: lister}
}

// List handles GET /users/{user_id}/items. The lookback window is a
// fixed server-side cutoff; clients cannot request a different range.
func (h *ItemsHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("user_id")
	// Empty user_id is unreachable via normal routing: ServeMux's path
	// cleaning redirects "/users//items" to "/users/items" (a 3-segment
	// pattern mismatch) before this handler ever runs. Kept as a guard
	// against future input paths (e.g. a differently-shaped route).
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
