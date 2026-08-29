package dynamo

import (
	"encoding/json/v2"
	"strings"
	"testing"
	"time"
)

func TestFormatSortKeyTimeFixedWidth(t *testing.T) {
	// ナノ秒がゼロの時刻でも、（time.RFC3339Nanoと違い）省略・空にならず
	// 9桁の表記を生成しなければならない。
	tm := time.Date(2026, 8, 29, 12, 0, 0, 0, time.UTC)

	got := formatSortKeyTime(tm)

	const want = "2026-08-29T12:00:00.000000000Z"
	if got != want {
		t.Fatalf("formatSortKeyTime(%v) = %q, want %q", tm, got, want)
	}
}

func TestFormatSortKeyTimePreservesChronologicalOrder(t *testing.T) {
	base := time.Date(2026, 8, 29, 12, 0, 0, 0, time.UTC)
	later := base.Add(100 * time.Nanosecond)

	baseStr := formatSortKeyTime(base)
	laterStr := formatSortKeyTime(later)

	if baseStr >= laterStr {
		t.Fatalf("lexicographic order broken: %q should be < %q", baseStr, laterStr)
	}
}

func TestComposeAndExtractSortKey(t *testing.T) {
	createdAt := formatSortKeyTime(time.Date(2026, 8, 29, 12, 0, 0, 0, time.UTC))
	jobID := "a1b2c3d4-5678-90ab-cdef-1234567890ab"

	sk := composeSortKey(createdAt, jobID)
	if !strings.HasPrefix(sk, createdAt+"#") {
		t.Fatalf("composeSortKey result %q does not start with %q#", sk, createdAt)
	}

	item := Item{SortKey: sk}
	if got := item.CreatedAt(); got != createdAt {
		t.Fatalf("Item.CreatedAt() = %q, want %q", got, createdAt)
	}
}

func TestItemMarshalJSONExcludesSortKey(t *testing.T) {
	createdAt := formatSortKeyTime(time.Date(2026, 8, 29, 12, 0, 0, 0, time.UTC))
	jobID := "a1b2c3d4-5678-90ab-cdef-1234567890ab"

	item := Item{
		UserID:  "user-1",
		SortKey: composeSortKey(createdAt, jobID),
		JobID:   jobID,
	}

	out, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(out, &decoded); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}

	if _, ok := decoded["created_at"]; !ok {
		t.Fatalf("expected \"created_at\" field, got %s", out)
	}
	if decoded["created_at"] != createdAt {
		t.Fatalf("created_at = %v, want %q", decoded["created_at"], createdAt)
	}
	if len(decoded) != 3 {
		t.Fatalf("expected exactly 3 fields (user_id, created_at, job_id), got %s", out)
	}
}
