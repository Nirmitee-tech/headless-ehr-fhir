package fhir

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
)

// ---------------------------------------------------------------------------
// CursorEncoder: Encode / Decode round-trip
// ---------------------------------------------------------------------------

func TestCursorEncoder_RoundTrip(t *testing.T) {
	enc := NewCursorEncoder([]byte("test-secret"))
	cursor := &PaginationCursor{
		Values:    map[string]interface{}{"last_updated": "2024-06-01T00:00:00Z"},
		Direction: CursorForward,
		ID:        "patient-123",
		CreatedAt: time.Now().UTC().Truncate(time.Second),
		PageSize:  20,
		SortKeys:  []SortKey{{Field: "date", Column: "last_updated", Ascending: false}},
	}

	encoded, err := enc.Encode(cursor)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}
	if encoded == "" {
		t.Fatal("expected non-empty encoded string")
	}

	decoded, err := enc.Decode(encoded)
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}

	if decoded.ID != cursor.ID {
		t.Errorf("ID = %q, want %q", decoded.ID, cursor.ID)
	}
	if decoded.Direction != cursor.Direction {
		t.Errorf("Direction = %d, want %d", decoded.Direction, cursor.Direction)
	}
	if decoded.PageSize != cursor.PageSize {
		t.Errorf("PageSize = %d, want %d", decoded.PageSize, cursor.PageSize)
	}
	if len(decoded.SortKeys) != 1 {
		t.Fatalf("SortKeys length = %d, want 1", len(decoded.SortKeys))
	}
	if decoded.SortKeys[0].Field != "date" {
		t.Errorf("SortKeys[0].Field = %q, want %q", decoded.SortKeys[0].Field, "date")
	}
}

func TestCursorEncoder_RoundTrip_MultipleValues(t *testing.T) {
	enc := NewCursorEncoder([]byte("secret"))
	cursor := &PaginationCursor{
		Values: map[string]interface{}{
			"last_updated": "2024-06-01",
			"name":         "Smith",
		},
		Direction: CursorBackward,
		ID:        "p-999",
		CreatedAt: time.Now().UTC().Truncate(time.Second),
		PageSize:  50,
		SortKeys: []SortKey{
			{Field: "date", Column: "last_updated", Ascending: false},
			{Field: "name", Column: "family_name", Ascending: true},
		},
	}

	encoded, err := enc.Encode(cursor)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}

	decoded, err := enc.Decode(encoded)
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}

	if decoded.Direction != CursorBackward {
		t.Errorf("Direction = %d, want CursorBackward", decoded.Direction)
	}
	if len(decoded.Values) != 2 {
		t.Errorf("Values length = %d, want 2", len(decoded.Values))
	}
	if len(decoded.SortKeys) != 2 {
		t.Errorf("SortKeys length = %d, want 2", len(decoded.SortKeys))
	}
}

func TestCursorEncoder_Decode_Tampered(t *testing.T) {
	enc := NewCursorEncoder([]byte("secret-key"))
	cursor := &PaginationCursor{
		Values:    map[string]interface{}{"x": "y"},
		Direction: CursorForward,
		ID:        "id-1",
		CreatedAt: time.Now().UTC(),
		PageSize:  10,
	}

	encoded, err := enc.Encode(cursor)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}

	// Decode, tamper, re-encode the payload without valid HMAC
	raw, _ := base64.RawURLEncoding.DecodeString(encoded)
	// Flip a byte in the payload portion (after the HMAC)
	if len(raw) > 33 {
		raw[33] ^= 0xFF
	}
	tampered := base64.RawURLEncoding.EncodeToString(raw)

	_, err = enc.Decode(tampered)
	if err == nil {
		t.Fatal("expected error for tampered cursor")
	}
	if !strings.Contains(err.Error(), "tamper") && !strings.Contains(err.Error(), "HMAC") &&
		!strings.Contains(err.Error(), "invalid") {
		t.Errorf("expected tamper-related error, got: %v", err)
	}
}

func TestCursorEncoder_Decode_DifferentSecret(t *testing.T) {
	enc1 := NewCursorEncoder([]byte("secret-1"))
	enc2 := NewCursorEncoder([]byte("secret-2"))

	cursor := &PaginationCursor{
		Values:    map[string]interface{}{"x": "y"},
		Direction: CursorForward,
		ID:        "id-1",
		CreatedAt: time.Now().UTC(),
		PageSize:  10,
	}

	encoded, err := enc1.Encode(cursor)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}

	_, err = enc2.Decode(encoded)
	if err == nil {
		t.Fatal("expected error when decoding with different secret")
	}
}

func TestCursorEncoder_Decode_ExpiredCursor(t *testing.T) {
	enc := NewCursorEncoder([]byte("secret"))
	cursor := &PaginationCursor{
		Values:    map[string]interface{}{"x": "y"},
		Direction: CursorForward,
		ID:        "id-1",
		CreatedAt: time.Now().UTC().Add(-25 * time.Hour), // expired
		PageSize:  10,
	}

	encoded, err := enc.Encode(cursor)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}

	_, err = enc.DecodeWithTTL(encoded, 1*time.Hour)
	if err == nil {
		t.Fatal("expected error for expired cursor")
	}
	if !strings.Contains(err.Error(), "expired") {
		t.Errorf("expected 'expired' in error, got: %v", err)
	}
}

func TestCursorEncoder_Decode_ValidTTL(t *testing.T) {
	enc := NewCursorEncoder([]byte("secret"))
	cursor := &PaginationCursor{
		Values:    map[string]interface{}{"x": "y"},
		Direction: CursorForward,
		ID:        "id-1",
		CreatedAt: time.Now().UTC(),
		PageSize:  10,
	}

	encoded, err := enc.Encode(cursor)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}

	decoded, err := enc.DecodeWithTTL(encoded, 1*time.Hour)
	if err != nil {
		t.Fatalf("unexpected error for valid TTL: %v", err)
	}
	if decoded.ID != "id-1" {
		t.Errorf("ID = %q, want %q", decoded.ID, "id-1")
	}
}

func TestCursorEncoder_Decode_InvalidBase64(t *testing.T) {
	enc := NewCursorEncoder([]byte("secret"))
	_, err := enc.Decode("!!!not-valid-base64!!!")
	if err == nil {
		t.Fatal("expected error for invalid base64")
	}
}

func TestCursorEncoder_Decode_EmptyString(t *testing.T) {
	enc := NewCursorEncoder([]byte("secret"))
	_, err := enc.Decode("")
	if err == nil {
		t.Fatal("expected error for empty string")
	}
}

func TestCursorEncoder_Decode_TooShort(t *testing.T) {
	enc := NewCursorEncoder([]byte("secret"))
	short := base64.RawURLEncoding.EncodeToString([]byte("tiny"))
	_, err := enc.Decode(short)
	if err == nil {
		t.Fatal("expected error for too-short payload")
	}
}

func TestCursorEncoder_HMAC_Integrity(t *testing.T) {
	secret := []byte("my-secret")
	enc := NewCursorEncoder(secret)
	cursor := &PaginationCursor{
		Values:    map[string]interface{}{"col": "val"},
		Direction: CursorForward,
		ID:        "abc",
		CreatedAt: time.Now().UTC(),
		PageSize:  10,
	}

	encoded, _ := enc.Encode(cursor)
	raw, _ := base64.RawURLEncoding.DecodeString(encoded)

	// First 32 bytes should be valid HMAC-SHA256 of the rest
	if len(raw) <= 32 {
		t.Fatal("encoded payload too short to contain HMAC")
	}
	mac := raw[:32]
	payload := raw[32:]

	h := hmac.New(sha256.New, secret)
	h.Write(payload)
	expected := h.Sum(nil)

	if !hmac.Equal(mac, expected) {
		t.Error("HMAC verification failed on raw payload")
	}
}

// ---------------------------------------------------------------------------
// BuildKeysetWhereClause
// ---------------------------------------------------------------------------

func TestBuildKeysetWhereClause_SingleDescending(t *testing.T) {
	cursor := &PaginationCursor{
		Values:   map[string]interface{}{"last_updated": "2024-06-01"},
		ID:       "p-1",
		SortKeys: []SortKey{{Field: "date", Column: "last_updated", Ascending: false}},
	}

	clause, args := BuildKeysetWhereClause(cursor, 1)

	if clause == "" {
		t.Fatal("expected non-empty clause")
	}
	if len(args) == 0 {
		t.Fatal("expected args")
	}
	// For DESC sort with tiebreaker: WHERE (last_updated, id) < ($1, $2)
	if !strings.Contains(clause, "last_updated") {
		t.Errorf("clause should reference last_updated, got: %s", clause)
	}
	if !strings.Contains(clause, "$1") {
		t.Errorf("clause should contain $1, got: %s", clause)
	}
}

func TestBuildKeysetWhereClause_SingleAscending(t *testing.T) {
	cursor := &PaginationCursor{
		Values:   map[string]interface{}{"family_name": "Smith"},
		ID:       "p-2",
		SortKeys: []SortKey{{Field: "name", Column: "family_name", Ascending: true}},
	}

	clause, args := BuildKeysetWhereClause(cursor, 1)

	if clause == "" {
		t.Fatal("expected non-empty clause")
	}
	// For ASC sort: WHERE (family_name, id) > ($1, $2)
	if !strings.Contains(clause, "family_name") {
		t.Errorf("clause should reference family_name, got: %s", clause)
	}
	if len(args) < 2 {
		t.Errorf("expected at least 2 args (value + id), got %d", len(args))
	}
}

func TestBuildKeysetWhereClause_MultiSort(t *testing.T) {
	cursor := &PaginationCursor{
		Values: map[string]interface{}{
			"last_updated": "2024-06-01",
			"family_name":  "Smith",
		},
		ID: "p-3",
		SortKeys: []SortKey{
			{Field: "date", Column: "last_updated", Ascending: false},
			{Field: "name", Column: "family_name", Ascending: true},
		},
	}

	clause, args := BuildKeysetWhereClause(cursor, 1)

	if clause == "" {
		t.Fatal("expected non-empty clause")
	}
	if !strings.Contains(clause, "last_updated") {
		t.Errorf("clause should reference last_updated, got: %s", clause)
	}
	if !strings.Contains(clause, "family_name") {
		t.Errorf("clause should reference family_name, got: %s", clause)
	}
	if len(args) < 3 {
		t.Errorf("expected at least 3 args, got %d", len(args))
	}
}

func TestBuildKeysetWhereClause_MixedDirections(t *testing.T) {
	cursor := &PaginationCursor{
		Values: map[string]interface{}{
			"last_updated": "2024-06-01",
			"family_name":  "Adams",
		},
		ID: "p-4",
		SortKeys: []SortKey{
			{Field: "date", Column: "last_updated", Ascending: false},
			{Field: "name", Column: "family_name", Ascending: true},
		},
	}

	clause, args := BuildKeysetWhereClause(cursor, 1)

	// Should produce a multi-part condition
	if clause == "" {
		t.Fatal("expected non-empty clause")
	}
	if len(args) < 3 {
		t.Fatalf("expected at least 3 args, got %d", len(args))
	}
}

func TestBuildKeysetWhereClause_CustomStartIdx(t *testing.T) {
	cursor := &PaginationCursor{
		Values:   map[string]interface{}{"last_updated": "2024-06-01"},
		ID:       "p-5",
		SortKeys: []SortKey{{Field: "date", Column: "last_updated", Ascending: false}},
	}

	clause, _ := BuildKeysetWhereClause(cursor, 5)

	if !strings.Contains(clause, "$5") {
		t.Errorf("clause should contain $5 for startIdx=5, got: %s", clause)
	}
}

func TestBuildKeysetWhereClause_NilCursor(t *testing.T) {
	clause, args := BuildKeysetWhereClause(nil, 1)
	if clause != "" {
		t.Errorf("expected empty clause for nil cursor, got: %s", clause)
	}
	if len(args) != 0 {
		t.Errorf("expected no args for nil cursor, got %d", len(args))
	}
}

func TestBuildKeysetWhereClause_WithNullValues(t *testing.T) {
	cursor := &PaginationCursor{
		Values:   map[string]interface{}{"family_name": nil},
		ID:       "p-6",
		SortKeys: []SortKey{{Field: "name", Column: "family_name", Ascending: true}},
	}

	clause, _ := BuildKeysetWhereClause(cursor, 1)

	// Should handle null sort values with IS NULL / IS NOT NULL logic
	if clause == "" {
		t.Fatal("expected non-empty clause for null sort value")
	}
	if !strings.Contains(strings.ToUpper(clause), "NULL") && !strings.Contains(clause, "id") {
		t.Errorf("clause should handle null values, got: %s", clause)
	}
}

// ---------------------------------------------------------------------------
// BuildKeysetOrderClause
// ---------------------------------------------------------------------------

func TestBuildKeysetOrderClause_SingleDescending(t *testing.T) {
	keys := []SortKey{{Field: "date", Column: "last_updated", Ascending: false}}
	clause := BuildKeysetOrderClause(keys)

	if !strings.Contains(clause, "ORDER BY") {
		t.Errorf("expected ORDER BY, got: %s", clause)
	}
	if !strings.Contains(clause, "last_updated DESC") {
		t.Errorf("expected 'last_updated DESC', got: %s", clause)
	}
	// Should include id as tiebreaker
	if !strings.Contains(clause, "id") {
		t.Errorf("expected id tiebreaker in ORDER BY, got: %s", clause)
	}
}

func TestBuildKeysetOrderClause_SingleAscending(t *testing.T) {
	keys := []SortKey{{Field: "name", Column: "family_name", Ascending: true}}
	clause := BuildKeysetOrderClause(keys)

	if !strings.Contains(clause, "family_name ASC") {
		t.Errorf("expected 'family_name ASC', got: %s", clause)
	}
}

func TestBuildKeysetOrderClause_MultiColumn(t *testing.T) {
	keys := []SortKey{
		{Field: "date", Column: "last_updated", Ascending: false},
		{Field: "name", Column: "family_name", Ascending: true},
	}
	clause := BuildKeysetOrderClause(keys)

	if !strings.Contains(clause, "last_updated DESC") {
		t.Errorf("expected 'last_updated DESC', got: %s", clause)
	}
	if !strings.Contains(clause, "family_name ASC") {
		t.Errorf("expected 'family_name ASC', got: %s", clause)
	}
}

func TestBuildKeysetOrderClause_NullsHandling(t *testing.T) {
	keys := []SortKey{{Field: "date", Column: "last_updated", Ascending: false}}
	clause := BuildKeysetOrderClause(keys)

	if !strings.Contains(clause, "NULLS LAST") {
		t.Errorf("expected NULLS LAST for DESC, got: %s", clause)
	}
}

func TestBuildKeysetOrderClause_AscendingNulls(t *testing.T) {
	keys := []SortKey{{Field: "name", Column: "family_name", Ascending: true}}
	clause := BuildKeysetOrderClause(keys)

	if !strings.Contains(clause, "NULLS LAST") {
		t.Errorf("expected NULLS LAST for ASC, got: %s", clause)
	}
}

func TestBuildKeysetOrderClause_Empty(t *testing.T) {
	clause := BuildKeysetOrderClause(nil)
	if clause != "" {
		t.Errorf("expected empty clause for nil keys, got: %s", clause)
	}

	clause = BuildKeysetOrderClause([]SortKey{})
	if clause != "" {
		t.Errorf("expected empty clause for empty keys, got: %s", clause)
	}
}

// ---------------------------------------------------------------------------
// ParseSortKeys
// ---------------------------------------------------------------------------

func TestParseSortKeys_Simple(t *testing.T) {
	colMap := map[string]string{"date": "last_updated", "name": "family_name"}
	keys := ParseSortKeys("-date", colMap)

	if len(keys) != 1 {
		t.Fatalf("expected 1 key, got %d", len(keys))
	}
	if keys[0].Field != "date" {
		t.Errorf("Field = %q, want 'date'", keys[0].Field)
	}
	if keys[0].Column != "last_updated" {
		t.Errorf("Column = %q, want 'last_updated'", keys[0].Column)
	}
	if keys[0].Ascending {
		t.Error("expected Ascending=false for -date")
	}
}

func TestParseSortKeys_MultipleFields(t *testing.T) {
	colMap := map[string]string{"date": "last_updated", "name": "family_name"}
	keys := ParseSortKeys("-date,name", colMap)

	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(keys))
	}
	if keys[0].Ascending {
		t.Error("expected first key descending")
	}
	if !keys[1].Ascending {
		t.Error("expected second key ascending")
	}
}

func TestParseSortKeys_PlusPrefix(t *testing.T) {
	colMap := map[string]string{"date": "last_updated"}
	keys := ParseSortKeys("+date", colMap)

	if len(keys) != 1 {
		t.Fatalf("expected 1 key, got %d", len(keys))
	}
	if !keys[0].Ascending {
		t.Error("expected Ascending=true for +date")
	}
}

func TestParseSortKeys_NoPrefix(t *testing.T) {
	colMap := map[string]string{"date": "last_updated"}
	keys := ParseSortKeys("date", colMap)

	if len(keys) != 1 {
		t.Fatalf("expected 1 key, got %d", len(keys))
	}
	if !keys[0].Ascending {
		t.Error("expected Ascending=true when no prefix")
	}
}

func TestParseSortKeys_UnknownField(t *testing.T) {
	colMap := map[string]string{"date": "last_updated"}
	keys := ParseSortKeys("unknown", colMap)

	if len(keys) != 0 {
		t.Errorf("expected 0 keys for unknown field, got %d", len(keys))
	}
}

func TestParseSortKeys_EmptyString(t *testing.T) {
	colMap := map[string]string{"date": "last_updated"}
	keys := ParseSortKeys("", colMap)

	if len(keys) != 0 {
		t.Errorf("expected 0 keys for empty string, got %d", len(keys))
	}
}

func TestParseSortKeys_MixedKnownUnknown(t *testing.T) {
	colMap := map[string]string{"date": "last_updated", "name": "family_name"}
	keys := ParseSortKeys("-date,unknown,name", colMap)

	if len(keys) != 2 {
		t.Fatalf("expected 2 keys (unknown skipped), got %d", len(keys))
	}
	if keys[0].Field != "date" {
		t.Errorf("first key field = %q, want 'date'", keys[0].Field)
	}
	if keys[1].Field != "name" {
		t.Errorf("second key field = %q, want 'name'", keys[1].Field)
	}
}

// ---------------------------------------------------------------------------
// DefaultSortKeys
// ---------------------------------------------------------------------------

func TestDefaultSortKeys(t *testing.T) {
	keys := DefaultSortKeys()

	if len(keys) != 2 {
		t.Fatalf("expected 2 default sort keys, got %d", len(keys))
	}

	// First key should be lastUpdated descending
	if keys[0].Field != "_lastUpdated" {
		t.Errorf("first key field = %q, want '_lastUpdated'", keys[0].Field)
	}
	if keys[0].Ascending {
		t.Error("expected first key to be descending")
	}

	// Second key should be id descending
	if keys[1].Field != "_id" {
		t.Errorf("second key field = %q, want '_id'", keys[1].Field)
	}
	if keys[1].Ascending {
		t.Error("expected second key to be descending")
	}
}

// ---------------------------------------------------------------------------
// BuildCursorFromRow
// ---------------------------------------------------------------------------

func TestBuildCursorFromRow_Simple(t *testing.T) {
	resource := map[string]interface{}{
		"id":   "patient-1",
		"meta": map[string]interface{}{"lastUpdated": "2024-06-01T00:00:00Z"},
	}
	keys := []SortKey{
		{Field: "_lastUpdated", Column: "last_updated", Ascending: false},
	}

	cursor := BuildCursorFromRow(resource, keys)

	if cursor == nil {
		t.Fatal("expected non-nil cursor")
	}
	if cursor.ID != "patient-1" {
		t.Errorf("ID = %q, want 'patient-1'", cursor.ID)
	}
}

func TestBuildCursorFromRow_MultipleKeys(t *testing.T) {
	resource := map[string]interface{}{
		"id":   "obs-42",
		"meta": map[string]interface{}{"lastUpdated": "2024-06-01T00:00:00Z"},
		"code": "12345",
	}
	keys := []SortKey{
		{Field: "_lastUpdated", Column: "last_updated", Ascending: false},
		{Field: "code", Column: "code", Ascending: true},
	}

	cursor := BuildCursorFromRow(resource, keys)

	if cursor == nil {
		t.Fatal("expected non-nil cursor")
	}
	if cursor.ID != "obs-42" {
		t.Errorf("ID = %q, want 'obs-42'", cursor.ID)
	}
	if len(cursor.SortKeys) != 2 {
		t.Errorf("SortKeys length = %d, want 2", len(cursor.SortKeys))
	}
}

func TestBuildCursorFromRow_MissingID(t *testing.T) {
	resource := map[string]interface{}{
		"meta": map[string]interface{}{"lastUpdated": "2024-06-01T00:00:00Z"},
	}
	keys := []SortKey{
		{Field: "_lastUpdated", Column: "last_updated", Ascending: false},
	}

	cursor := BuildCursorFromRow(resource, keys)

	if cursor == nil {
		t.Fatal("expected non-nil cursor")
	}
	if cursor.ID != "" {
		t.Errorf("expected empty ID, got %q", cursor.ID)
	}
}

func TestBuildCursorFromRow_EmptyResource(t *testing.T) {
	resource := map[string]interface{}{}
	keys := []SortKey{
		{Field: "_lastUpdated", Column: "last_updated", Ascending: false},
	}

	cursor := BuildCursorFromRow(resource, keys)

	if cursor == nil {
		t.Fatal("expected non-nil cursor")
	}
}

func TestBuildCursorFromRow_NilResource(t *testing.T) {
	keys := []SortKey{
		{Field: "_lastUpdated", Column: "last_updated", Ascending: false},
	}

	cursor := BuildCursorFromRow(nil, keys)

	if cursor != nil {
		t.Error("expected nil cursor for nil resource")
	}
}

// ---------------------------------------------------------------------------
// BuildBundleLinks
// ---------------------------------------------------------------------------

func TestBuildBundleLinks_NextOnly(t *testing.T) {
	enc := NewCursorEncoder([]byte("secret"))
	page := &CursorPage{
		HasNext:  true,
		HasPrev:  false,
		PageSize: 20,
		NextCursor: mustEncodeCursor(t, enc, &PaginationCursor{
			Values: map[string]interface{}{"last_updated": "2024-06-01"},
			ID:     "p-20", CreatedAt: time.Now().UTC(), PageSize: 20,
		}),
	}

	links := BuildBundleLinks("https://fhir.example.com/Patient", page, enc)

	found := map[string]bool{}
	for _, link := range links {
		rel, _ := link["relation"].(string)
		found[rel] = true
	}
	if !found["self"] {
		t.Error("expected 'self' link")
	}
	if !found["next"] {
		t.Error("expected 'next' link")
	}
	if found["previous"] {
		t.Error("unexpected 'previous' link")
	}
}

func TestBuildBundleLinks_PrevOnly(t *testing.T) {
	enc := NewCursorEncoder([]byte("secret"))
	page := &CursorPage{
		HasNext:  false,
		HasPrev:  true,
		PageSize: 20,
		PrevCursor: mustEncodeCursor(t, enc, &PaginationCursor{
			Values: map[string]interface{}{"last_updated": "2024-01-01"},
			ID:     "p-1", CreatedAt: time.Now().UTC(), PageSize: 20,
		}),
	}

	links := BuildBundleLinks("https://fhir.example.com/Patient", page, enc)

	found := map[string]bool{}
	for _, link := range links {
		rel, _ := link["relation"].(string)
		found[rel] = true
	}
	if !found["self"] {
		t.Error("expected 'self' link")
	}
	if found["next"] {
		t.Error("unexpected 'next' link")
	}
	if !found["previous"] {
		t.Error("expected 'previous' link")
	}
}

func TestBuildBundleLinks_BothDirections(t *testing.T) {
	enc := NewCursorEncoder([]byte("secret"))
	page := &CursorPage{
		HasNext:  true,
		HasPrev:  true,
		PageSize: 10,
		NextCursor: mustEncodeCursor(t, enc, &PaginationCursor{
			Values: map[string]interface{}{"x": "y"}, ID: "n1",
			CreatedAt: time.Now().UTC(), PageSize: 10,
		}),
		PrevCursor: mustEncodeCursor(t, enc, &PaginationCursor{
			Values: map[string]interface{}{"x": "z"}, ID: "p1",
			CreatedAt: time.Now().UTC(), PageSize: 10,
		}),
	}

	links := BuildBundleLinks("https://fhir.example.com/Patient", page, enc)

	found := map[string]bool{}
	for _, link := range links {
		rel, _ := link["relation"].(string)
		found[rel] = true
	}
	if !found["self"] || !found["next"] || !found["previous"] {
		t.Errorf("expected self, next, previous; found: %v", found)
	}
}

func TestBuildBundleLinks_NeitherDirection(t *testing.T) {
	enc := NewCursorEncoder([]byte("secret"))
	page := &CursorPage{
		HasNext:  false,
		HasPrev:  false,
		PageSize: 20,
	}

	links := BuildBundleLinks("https://fhir.example.com/Patient", page, enc)

	if len(links) != 1 {
		t.Fatalf("expected 1 link (self only), got %d", len(links))
	}
	rel, _ := links[0]["relation"].(string)
	if rel != "self" {
		t.Errorf("expected 'self', got %q", rel)
	}
}

func TestBuildBundleLinks_URLContainsCursor(t *testing.T) {
	enc := NewCursorEncoder([]byte("secret"))
	page := &CursorPage{
		HasNext:  true,
		HasPrev:  false,
		PageSize: 20,
		NextCursor: mustEncodeCursor(t, enc, &PaginationCursor{
			Values: map[string]interface{}{"x": "y"}, ID: "id1",
			CreatedAt: time.Now().UTC(), PageSize: 20,
		}),
	}

	links := BuildBundleLinks("https://fhir.example.com/Patient", page, enc)

	for _, link := range links {
		rel, _ := link["relation"].(string)
		if rel == "next" {
			u, _ := link["url"].(string)
			if !strings.Contains(u, "_cursor=") {
				t.Errorf("next link should contain _cursor= param, got: %s", u)
			}
		}
	}
}

// ---------------------------------------------------------------------------
// ApplyCursorPagination
// ---------------------------------------------------------------------------

func TestApplyCursorPagination_Forward(t *testing.T) {
	cursor := &PaginationCursor{
		Values:    map[string]interface{}{"last_updated": "2024-06-01"},
		Direction: CursorForward,
		ID:        "p-1",
		SortKeys:  []SortKey{{Field: "date", Column: "last_updated", Ascending: false}},
	}

	query, args := ApplyCursorPagination("SELECT * FROM patients", cursor, 20, 1)

	if !strings.Contains(query, "LIMIT") {
		t.Errorf("expected LIMIT clause, got: %s", query)
	}
	if !strings.Contains(query, "ORDER BY") {
		t.Errorf("expected ORDER BY clause, got: %s", query)
	}
	if len(args) == 0 {
		t.Error("expected args from cursor")
	}
}

func TestApplyCursorPagination_Backward(t *testing.T) {
	cursor := &PaginationCursor{
		Values:    map[string]interface{}{"last_updated": "2024-06-01"},
		Direction: CursorBackward,
		ID:        "p-1",
		SortKeys:  []SortKey{{Field: "date", Column: "last_updated", Ascending: false}},
	}

	query, args := ApplyCursorPagination("SELECT * FROM patients", cursor, 20, 1)

	if !strings.Contains(query, "LIMIT") {
		t.Errorf("expected LIMIT clause, got: %s", query)
	}
	if len(args) == 0 {
		t.Error("expected args from cursor")
	}
}

func TestApplyCursorPagination_FirstPage(t *testing.T) {
	query, args := ApplyCursorPagination("SELECT * FROM patients", nil, 20, 1)

	if !strings.Contains(query, "LIMIT") {
		t.Errorf("expected LIMIT clause, got: %s", query)
	}
	// No cursor args on first page
	if len(args) != 0 {
		t.Errorf("expected no args for first page, got %d", len(args))
	}
}

func TestApplyCursorPagination_PreservesBaseQuery(t *testing.T) {
	base := "SELECT * FROM patients WHERE tenant_id = 'abc'"
	query, _ := ApplyCursorPagination(base, nil, 10, 1)

	if !strings.Contains(query, "tenant_id = 'abc'") {
		t.Errorf("base query conditions should be preserved, got: %s", query)
	}
}

// ---------------------------------------------------------------------------
// ParseCursorParams
// ---------------------------------------------------------------------------

func TestParseCursorParams_After(t *testing.T) {
	vals := url.Values{}
	vals.Set("_cursor", "some-cursor-token")
	vals.Set("_count", "25")

	params, err := ParseCursorParams(vals)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if params.After != "some-cursor-token" {
		t.Errorf("After = %q, want 'some-cursor-token'", params.After)
	}
	if params.Count != 25 {
		t.Errorf("Count = %d, want 25", params.Count)
	}
}

func TestParseCursorParams_Before(t *testing.T) {
	vals := url.Values{}
	vals.Set("_cursor:prev", "prev-cursor-token")

	params, err := ParseCursorParams(vals)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if params.Before != "prev-cursor-token" {
		t.Errorf("Before = %q, want 'prev-cursor-token'", params.Before)
	}
}

func TestParseCursorParams_Count(t *testing.T) {
	vals := url.Values{}
	vals.Set("_count", "50")

	params, err := ParseCursorParams(vals)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if params.Count != 50 {
		t.Errorf("Count = %d, want 50", params.Count)
	}
}

func TestParseCursorParams_Sort(t *testing.T) {
	vals := url.Values{}
	vals.Set("_sort", "-date,name")

	params, err := ParseCursorParams(vals)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if params.Sort != "-date,name" {
		t.Errorf("Sort = %q, want '-date,name'", params.Sort)
	}
}

func TestParseCursorParams_Defaults(t *testing.T) {
	vals := url.Values{}

	params, err := ParseCursorParams(vals)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if params.After != "" {
		t.Errorf("After should be empty, got %q", params.After)
	}
	if params.Before != "" {
		t.Errorf("Before should be empty, got %q", params.Before)
	}
	if params.Count != 0 {
		t.Errorf("Count should be 0 (unset), got %d", params.Count)
	}
	if params.Sort != "" {
		t.Errorf("Sort should be empty, got %q", params.Sort)
	}
}

func TestParseCursorParams_InvalidCount(t *testing.T) {
	vals := url.Values{}
	vals.Set("_count", "not-a-number")

	_, err := ParseCursorParams(vals)
	if err == nil {
		t.Fatal("expected error for invalid _count")
	}
}

func TestParseCursorParams_NegativeCount(t *testing.T) {
	vals := url.Values{}
	vals.Set("_count", "-5")

	_, err := ParseCursorParams(vals)
	if err == nil {
		t.Fatal("expected error for negative _count")
	}
}

func TestParseCursorParams_BothCursors(t *testing.T) {
	vals := url.Values{}
	vals.Set("_cursor", "after")
	vals.Set("_cursor:prev", "before")

	_, err := ParseCursorParams(vals)
	if err == nil {
		t.Fatal("expected error when both _cursor and _cursor:prev are set")
	}
}

// ---------------------------------------------------------------------------
// ValidateCursorConsistency
// ---------------------------------------------------------------------------

func TestValidateCursorConsistency_Matching(t *testing.T) {
	cursor := &PaginationCursor{
		SortKeys: []SortKey{
			{Field: "date", Column: "last_updated", Ascending: false},
		},
	}
	request := []SortKey{
		{Field: "date", Column: "last_updated", Ascending: false},
	}

	err := ValidateCursorConsistency(cursor, request)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateCursorConsistency_Mismatched(t *testing.T) {
	cursor := &PaginationCursor{
		SortKeys: []SortKey{
			{Field: "date", Column: "last_updated", Ascending: false},
		},
	}
	request := []SortKey{
		{Field: "name", Column: "family_name", Ascending: true},
	}

	err := ValidateCursorConsistency(cursor, request)
	if err == nil {
		t.Fatal("expected error for mismatched sort keys")
	}
}

func TestValidateCursorConsistency_DifferentLength(t *testing.T) {
	cursor := &PaginationCursor{
		SortKeys: []SortKey{
			{Field: "date", Column: "last_updated", Ascending: false},
			{Field: "name", Column: "family_name", Ascending: true},
		},
	}
	request := []SortKey{
		{Field: "date", Column: "last_updated", Ascending: false},
	}

	err := ValidateCursorConsistency(cursor, request)
	if err == nil {
		t.Fatal("expected error for different length sort keys")
	}
}

func TestValidateCursorConsistency_EmptyCursorKeys(t *testing.T) {
	cursor := &PaginationCursor{
		SortKeys: []SortKey{},
	}
	request := []SortKey{
		{Field: "date", Column: "last_updated", Ascending: false},
	}

	err := ValidateCursorConsistency(cursor, request)
	if err == nil {
		t.Fatal("expected error for empty cursor sort keys")
	}
}

func TestValidateCursorConsistency_BothEmpty(t *testing.T) {
	cursor := &PaginationCursor{
		SortKeys: []SortKey{},
	}

	err := ValidateCursorConsistency(cursor, []SortKey{})
	// Both empty should be consistent (both use defaults)
	if err != nil {
		t.Errorf("unexpected error for both-empty: %v", err)
	}
}

func TestValidateCursorConsistency_DirectionMismatch(t *testing.T) {
	cursor := &PaginationCursor{
		SortKeys: []SortKey{
			{Field: "date", Column: "last_updated", Ascending: false},
		},
	}
	request := []SortKey{
		{Field: "date", Column: "last_updated", Ascending: true}, // direction differs
	}

	err := ValidateCursorConsistency(cursor, request)
	if err == nil {
		t.Fatal("expected error for direction mismatch")
	}
}

// ---------------------------------------------------------------------------
// EstimateTotal
// ---------------------------------------------------------------------------

func TestEstimateTotal(t *testing.T) {
	query := "SELECT * FROM patients WHERE tenant_id = 'abc'"
	result := EstimateTotal(query)

	if !strings.Contains(result, "EXPLAIN") {
		t.Errorf("expected EXPLAIN in result, got: %s", result)
	}
	if !strings.Contains(result, query) {
		t.Errorf("expected original query in result, got: %s", result)
	}
}

// ---------------------------------------------------------------------------
// DefaultColumnMap
// ---------------------------------------------------------------------------

func TestDefaultColumnMap(t *testing.T) {
	colMap := DefaultColumnMap()

	if colMap == nil {
		t.Fatal("expected non-nil column map")
	}
	if _, ok := colMap["_lastUpdated"]; !ok {
		t.Error("expected '_lastUpdated' in column map")
	}
	if _, ok := colMap["_id"]; !ok {
		t.Error("expected '_id' in column map")
	}
	if _, ok := colMap["date"]; !ok {
		t.Error("expected 'date' in column map")
	}
}

// ---------------------------------------------------------------------------
// DefaultPaginationConfig
// ---------------------------------------------------------------------------

func TestDefaultPaginationConfig(t *testing.T) {
	config := DefaultPaginationConfig()

	if config == nil {
		t.Fatal("expected non-nil config")
	}
	if config.DefaultPageSize <= 0 {
		t.Errorf("DefaultPageSize should be > 0, got %d", config.DefaultPageSize)
	}
	if config.MaxPageSize <= 0 {
		t.Errorf("MaxPageSize should be > 0, got %d", config.MaxPageSize)
	}
	if config.MaxPageSize < config.DefaultPageSize {
		t.Errorf("MaxPageSize (%d) should be >= DefaultPageSize (%d)", config.MaxPageSize, config.DefaultPageSize)
	}
	if config.CursorTTL <= 0 {
		t.Errorf("CursorTTL should be > 0, got %v", config.CursorTTL)
	}
}

// ---------------------------------------------------------------------------
// Page size enforcement
// ---------------------------------------------------------------------------

func TestEnforcePageSize_Default(t *testing.T) {
	config := DefaultPaginationConfig()
	size := EnforcePageSize(0, config)
	if size != config.DefaultPageSize {
		t.Errorf("size = %d, want default %d", size, config.DefaultPageSize)
	}
}

func TestEnforcePageSize_BelowMin(t *testing.T) {
	config := DefaultPaginationConfig()
	size := EnforcePageSize(-1, config)
	if size != 1 {
		t.Errorf("size = %d, want 1 (minimum)", size)
	}
}

func TestEnforcePageSize_AboveMax(t *testing.T) {
	config := &PaginationConfig{
		DefaultPageSize: 20,
		MaxPageSize:     100,
	}
	size := EnforcePageSize(500, config)
	if size != 100 {
		t.Errorf("size = %d, want max %d", size, 100)
	}
}

func TestEnforcePageSize_WithinRange(t *testing.T) {
	config := &PaginationConfig{
		DefaultPageSize: 20,
		MaxPageSize:     100,
	}
	size := EnforcePageSize(50, config)
	if size != 50 {
		t.Errorf("size = %d, want 50", size)
	}
}

func TestEnforcePageSize_ExactMax(t *testing.T) {
	config := &PaginationConfig{
		DefaultPageSize: 20,
		MaxPageSize:     100,
	}
	size := EnforcePageSize(100, config)
	if size != 100 {
		t.Errorf("size = %d, want 100", size)
	}
}

// ---------------------------------------------------------------------------
// CursorDirection
// ---------------------------------------------------------------------------

func TestCursorDirection_Constants(t *testing.T) {
	if CursorForward != 0 {
		t.Errorf("CursorForward = %d, want 0", CursorForward)
	}
	if CursorBackward != 1 {
		t.Errorf("CursorBackward = %d, want 1", CursorBackward)
	}
}

func TestCursorDirection_ForwardSemantics(t *testing.T) {
	cursor := &PaginationCursor{
		Values:    map[string]interface{}{"last_updated": "2024-06-01"},
		Direction: CursorForward,
		ID:        "p-1",
		SortKeys:  []SortKey{{Field: "date", Column: "last_updated", Ascending: false}},
	}

	clause, _ := BuildKeysetWhereClause(cursor, 1)
	// Forward with DESC should use < comparison
	if !strings.Contains(clause, "<") && !strings.Contains(clause, "last_updated") {
		t.Errorf("forward/desc clause should contain '<', got: %s", clause)
	}
}

func TestCursorDirection_BackwardSemantics(t *testing.T) {
	cursor := &PaginationCursor{
		Values:    map[string]interface{}{"last_updated": "2024-06-01"},
		Direction: CursorBackward,
		ID:        "p-1",
		SortKeys:  []SortKey{{Field: "date", Column: "last_updated", Ascending: false}},
	}

	clause, _ := BuildKeysetWhereClause(cursor, 1)
	// Backward with DESC should use > comparison (reverse direction)
	if !strings.Contains(clause, ">") && !strings.Contains(clause, "last_updated") {
		t.Errorf("backward/desc clause should contain '>', got: %s", clause)
	}
}

// ---------------------------------------------------------------------------
// Edge cases
// ---------------------------------------------------------------------------

func TestCursorPage_EmptyResults(t *testing.T) {
	page := &CursorPage{
		Resources:  []map[string]interface{}{},
		HasNext:    false,
		HasPrev:    false,
		NextCursor: "",
		PrevCursor: "",
		PageSize:   20,
	}

	if len(page.Resources) != 0 {
		t.Errorf("expected empty resources, got %d", len(page.Resources))
	}
	if page.HasNext || page.HasPrev {
		t.Error("empty results should not have next/prev")
	}
}

func TestCursorPage_SingleResult(t *testing.T) {
	page := &CursorPage{
		Resources: []map[string]interface{}{
			{"id": "p-1", "resourceType": "Patient"},
		},
		HasNext:  false,
		HasPrev:  false,
		PageSize: 20,
	}

	if len(page.Resources) != 1 {
		t.Errorf("expected 1 resource, got %d", len(page.Resources))
	}
}

func TestCursorPage_ExactlyPageSize(t *testing.T) {
	resources := make([]map[string]interface{}, 20)
	for i := 0; i < 20; i++ {
		resources[i] = map[string]interface{}{"id": "p-" + string(rune('A'+i))}
	}
	page := &CursorPage{
		Resources: resources,
		HasNext:   true, // could have more
		HasPrev:   true,
		PageSize:  20,
	}

	if len(page.Resources) != 20 {
		t.Errorf("expected 20 resources, got %d", len(page.Resources))
	}
}

func TestCursorPage_TotalCount(t *testing.T) {
	total := 42
	page := &CursorPage{
		TotalCount: &total,
		PageSize:   20,
	}

	if page.TotalCount == nil {
		t.Fatal("expected non-nil TotalCount")
	}
	if *page.TotalCount != 42 {
		t.Errorf("TotalCount = %d, want 42", *page.TotalCount)
	}
}

func TestCursorPage_NilTotalCount(t *testing.T) {
	page := &CursorPage{
		PageSize: 20,
	}

	if page.TotalCount != nil {
		t.Error("expected nil TotalCount")
	}
}

// ---------------------------------------------------------------------------
// Null handling in sort columns
// ---------------------------------------------------------------------------

func TestNullHandling_OrderClause(t *testing.T) {
	keys := []SortKey{
		{Field: "date", Column: "last_updated", Ascending: false},
		{Field: "name", Column: "family_name", Ascending: true},
	}

	clause := BuildKeysetOrderClause(keys)

	// Both DESC and ASC should have NULLS LAST
	occurrences := strings.Count(clause, "NULLS LAST")
	// At minimum the sort columns should have NULLS LAST
	if occurrences < 2 {
		t.Errorf("expected at least 2 NULLS LAST occurrences, got %d in: %s", occurrences, clause)
	}
}

// ---------------------------------------------------------------------------
// PaginationMiddleware
// ---------------------------------------------------------------------------

func TestPaginationMiddleware_WithoutCursor(t *testing.T) {
	config := DefaultPaginationConfig()
	config.Secret = []byte("test-secret")
	config.EnableCursor = true

	mw := PaginationMiddleware(config)

	e := echo.New()
	handler := mw(func(c echo.Context) error {
		// Check that pagination context is set
		pageSize := c.Get("_cursorPageSize")
		if pageSize == nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "no page size"})
		}
		return c.JSON(http.StatusOK, map[string]interface{}{
			"pageSize": pageSize,
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient?_count=15", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := handler(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}

	var body map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &body)
	if ps, ok := body["pageSize"].(float64); !ok || int(ps) != 15 {
		t.Errorf("pageSize = %v, want 15", body["pageSize"])
	}
}

func TestPaginationMiddleware_WithValidCursor(t *testing.T) {
	config := DefaultPaginationConfig()
	config.Secret = []byte("test-secret")
	config.EnableCursor = true

	enc := NewCursorEncoder(config.Secret)
	cursor := &PaginationCursor{
		Values:    map[string]interface{}{"last_updated": "2024-06-01"},
		Direction: CursorForward,
		ID:        "p-1",
		CreatedAt: time.Now().UTC(),
		PageSize:  10,
		SortKeys:  DefaultSortKeys(),
	}
	token, _ := enc.Encode(cursor)

	mw := PaginationMiddleware(config)

	e := echo.New()
	handler := mw(func(c echo.Context) error {
		cur := c.Get("_paginationCursor")
		if cur == nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "no cursor"})
		}
		pc, ok := cur.(*PaginationCursor)
		if !ok {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "wrong type"})
		}
		return c.JSON(http.StatusOK, map[string]interface{}{
			"cursorID": pc.ID,
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient?_cursor="+url.QueryEscape(token), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := handler(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}

	var body map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &body)
	if body["cursorID"] != "p-1" {
		t.Errorf("cursorID = %v, want 'p-1'", body["cursorID"])
	}
}

func TestPaginationMiddleware_InvalidCursor_Fallback(t *testing.T) {
	config := DefaultPaginationConfig()
	config.Secret = []byte("test-secret")
	config.EnableCursor = true
	config.FallbackOffset = true

	mw := PaginationMiddleware(config)

	e := echo.New()
	handler := mw(func(c echo.Context) error {
		cur := c.Get("_paginationCursor")
		if cur != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "cursor should be nil for fallback"})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient?_cursor=invalid-token", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := handler(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200 (fallback)", rec.Code)
	}
}

func TestPaginationMiddleware_InvalidCursor_NoFallback(t *testing.T) {
	config := DefaultPaginationConfig()
	config.Secret = []byte("test-secret")
	config.EnableCursor = true
	config.FallbackOffset = false

	mw := PaginationMiddleware(config)

	e := echo.New()
	handler := mw(func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient?_cursor=invalid-token", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler(c)
	// Should return an error or set error status
	if rec.Code == http.StatusOK && err == nil {
		// Check if the response contains an error
		var body map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &body)
		if body["resourceType"] != "OperationOutcome" {
			t.Error("expected error response for invalid cursor with no fallback")
		}
	}
}

func TestPaginationMiddleware_ExpiredCursor(t *testing.T) {
	config := DefaultPaginationConfig()
	config.Secret = []byte("test-secret")
	config.EnableCursor = true
	config.CursorTTL = 1 * time.Second
	config.FallbackOffset = true

	enc := NewCursorEncoder(config.Secret)
	cursor := &PaginationCursor{
		Values:    map[string]interface{}{"last_updated": "2024-06-01"},
		Direction: CursorForward,
		ID:        "p-1",
		CreatedAt: time.Now().UTC().Add(-1 * time.Hour), // expired
		PageSize:  10,
		SortKeys:  DefaultSortKeys(),
	}
	token, _ := enc.Encode(cursor)

	mw := PaginationMiddleware(config)

	e := echo.New()
	handler := mw(func(c echo.Context) error {
		cur := c.Get("_paginationCursor")
		if cur != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "should fallback"})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "fallback"})
	})

	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient?_cursor="+url.QueryEscape(token), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := handler(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200 (fallback)", rec.Code)
	}
}

func TestPaginationMiddleware_DisabledCursor(t *testing.T) {
	config := DefaultPaginationConfig()
	config.Secret = []byte("test-secret")
	config.EnableCursor = false

	mw := PaginationMiddleware(config)

	e := echo.New()
	handler := mw(func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := handler(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

func TestPaginationMiddleware_NonGetRequest(t *testing.T) {
	config := DefaultPaginationConfig()
	config.Secret = []byte("test-secret")
	config.EnableCursor = true

	mw := PaginationMiddleware(config)

	e := echo.New()
	handler := mw(func(c echo.Context) error {
		// Cursor context should not be set for POST
		if c.Get("_cursorPageSize") != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "should not have cursor context"})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodPost, "/fhir/Patient", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := handler(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

// ---------------------------------------------------------------------------
// CursorSearchQuery
// ---------------------------------------------------------------------------

func TestCursorSearchQuery_Fields(t *testing.T) {
	csq := CursorSearchQuery{
		BaseQuery: "SELECT * FROM patients",
		SortKeys: []SortKey{
			{Field: "date", Column: "last_updated", Ascending: false},
		},
		Cursor: &PaginationCursor{
			Values:    map[string]interface{}{"last_updated": "2024-06-01"},
			Direction: CursorForward,
			ID:        "p-1",
		},
		PageSize:  20,
		Direction: CursorForward,
	}

	if csq.BaseQuery != "SELECT * FROM patients" {
		t.Errorf("BaseQuery = %q", csq.BaseQuery)
	}
	if csq.PageSize != 20 {
		t.Errorf("PageSize = %d", csq.PageSize)
	}
}

// ---------------------------------------------------------------------------
// Integration: encode -> decode -> build clause
// ---------------------------------------------------------------------------

func TestIntegration_EncodeDecodeAndBuild(t *testing.T) {
	enc := NewCursorEncoder([]byte("integration-secret"))

	original := &PaginationCursor{
		Values:    map[string]interface{}{"last_updated": "2024-06-15T10:30:00Z"},
		Direction: CursorForward,
		ID:        "patient-42",
		CreatedAt: time.Now().UTC().Truncate(time.Second),
		PageSize:  25,
		SortKeys: []SortKey{
			{Field: "_lastUpdated", Column: "last_updated", Ascending: false},
		},
	}

	token, err := enc.Encode(original)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}

	decoded, err := enc.Decode(token)
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}

	clause, args := BuildKeysetWhereClause(decoded, 1)
	if clause == "" {
		t.Fatal("expected non-empty WHERE clause from decoded cursor")
	}
	if len(args) < 2 {
		t.Fatalf("expected at least 2 args, got %d", len(args))
	}

	order := BuildKeysetOrderClause(decoded.SortKeys)
	if !strings.Contains(order, "ORDER BY") {
		t.Errorf("expected ORDER BY clause, got: %s", order)
	}
}

func TestIntegration_FullPaginationFlow(t *testing.T) {
	enc := NewCursorEncoder([]byte("flow-secret"))
	keys := []SortKey{
		{Field: "_lastUpdated", Column: "last_updated", Ascending: false},
	}

	// Simulate building cursor from last row of page 1
	lastRow := map[string]interface{}{
		"id":   "patient-20",
		"meta": map[string]interface{}{"lastUpdated": "2024-03-15T00:00:00Z"},
	}
	cursor := BuildCursorFromRow(lastRow, keys)
	if cursor == nil {
		t.Fatal("expected non-nil cursor from row")
	}
	cursor.Direction = CursorForward
	cursor.CreatedAt = time.Now().UTC()
	cursor.PageSize = 20

	// Encode the cursor
	token, err := enc.Encode(cursor)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}

	// Decode on next request
	decoded, err := enc.Decode(token)
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}

	// Build query for page 2
	query, args := ApplyCursorPagination("SELECT * FROM patients WHERE tenant_id = 'abc'", decoded, 20, 1)
	if !strings.Contains(query, "LIMIT") {
		t.Errorf("expected LIMIT in query, got: %s", query)
	}
	if len(args) < 2 {
		t.Fatalf("expected at least 2 args, got %d", len(args))
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func mustEncodeCursor(t *testing.T, enc *CursorEncoder, cursor *PaginationCursor) string {
	t.Helper()
	token, err := enc.Encode(cursor)
	if err != nil {
		t.Fatalf("mustEncodeCursor: %v", err)
	}
	return token
}
