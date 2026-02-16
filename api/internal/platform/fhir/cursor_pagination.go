package fhir

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

// CursorDirection indicates forward or backward pagination.
type CursorDirection int

const (
	// CursorForward paginates forward (next page).
	CursorForward CursorDirection = iota
	// CursorBackward paginates backward (previous page).
	CursorBackward
)

// PaginationCursor represents an opaque cursor for keyset pagination.
// It captures the sort key values at a specific position in the result set,
// enabling efficient page traversal without OFFSET.
type PaginationCursor struct {
	Values    map[string]interface{} `json:"v"`  // Sort key values at the cursor position
	Direction CursorDirection        `json:"d"`  // Forward or backward
	ID        string                 `json:"id"` // Resource ID at cursor position (tiebreaker)
	CreatedAt time.Time              `json:"c"`  // When this cursor was created
	PageSize  int                    `json:"ps"` // Requested page size
	SortKeys  []SortKey              `json:"sk"` // The sort order used
}

// SortKey represents a single sort field and direction.
type SortKey struct {
	Field     string `json:"f"`  // FHIR parameter name
	Column    string `json:"co"` // SQL column name
	Ascending bool   `json:"a"`  // Sort direction
}

// CursorPage represents a page of results with cursor information.
type CursorPage struct {
	Resources  []map[string]interface{} // The resources in this page
	NextCursor string                   // Encoded cursor for next page (empty if no more)
	PrevCursor string                   // Encoded cursor for previous page (empty if first)
	TotalCount *int                     // Optional total count (only if requested)
	HasNext    bool                     // Whether there are more results after this page
	HasPrev    bool                     // Whether there are results before this page
	PageSize   int                      // The page size used
}

// CursorEncoder handles encoding/decoding of pagination cursors with HMAC signing.
type CursorEncoder struct {
	Secret []byte
}

// PaginationConfig configures pagination behavior.
type PaginationConfig struct {
	DefaultPageSize int           // Default page size when _count is not specified
	MaxPageSize     int           // Maximum allowed page size
	CursorTTL       time.Duration // How long cursors remain valid
	EnableCursor    bool          // Whether to use cursor-based pagination
	FallbackOffset  bool          // Fall back to offset if cursor invalid
	Secret          []byte        // HMAC signing key
}

// CursorSearchQuery extends search with cursor support.
type CursorSearchQuery struct {
	BaseQuery string
	SortKeys  []SortKey
	Cursor    *PaginationCursor
	PageSize  int
	Direction CursorDirection
}

// CursorParams holds raw pagination parameters extracted from a request.
type CursorParams struct {
	After  string // cursor for next page (_cursor)
	Before string // cursor for previous page (_cursor:prev)
	Count  int    // page size (_count)
	Sort   string // sort specification (_sort)
}

// NewCursorEncoder creates an encoder with HMAC signing for tamper detection.
func NewCursorEncoder(secret []byte) *CursorEncoder {
	return &CursorEncoder{Secret: secret}
}

// Encode serializes a PaginationCursor to an opaque base64 string with HMAC.
// The format is: base64(HMAC(32 bytes) + JSON payload).
func (e *CursorEncoder) Encode(cursor *PaginationCursor) (string, error) {
	payload, err := json.Marshal(cursor)
	if err != nil {
		return "", fmt.Errorf("failed to marshal cursor: %w", err)
	}

	mac := hmac.New(sha256.New, e.Secret)
	mac.Write(payload)
	sig := mac.Sum(nil) // 32 bytes

	// sig + payload
	combined := make([]byte, len(sig)+len(payload))
	copy(combined, sig)
	copy(combined[len(sig):], payload)

	return base64.RawURLEncoding.EncodeToString(combined), nil
}

// Decode deserializes an opaque cursor string back to PaginationCursor.
// It verifies the HMAC signature to detect tampering.
func (e *CursorEncoder) Decode(encoded string) (*PaginationCursor, error) {
	if encoded == "" {
		return nil, fmt.Errorf("invalid cursor: empty string")
	}

	raw, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("invalid cursor: bad base64: %w", err)
	}

	if len(raw) <= 32 {
		return nil, fmt.Errorf("invalid cursor: payload too short")
	}

	sig := raw[:32]
	payload := raw[32:]

	// Verify HMAC
	mac := hmac.New(sha256.New, e.Secret)
	mac.Write(payload)
	expected := mac.Sum(nil)

	if !hmac.Equal(sig, expected) {
		return nil, fmt.Errorf("invalid cursor: HMAC verification failed (tampered or wrong key)")
	}

	var cursor PaginationCursor
	if err := json.Unmarshal(payload, &cursor); err != nil {
		return nil, fmt.Errorf("invalid cursor: bad JSON payload: %w", err)
	}

	return &cursor, nil
}

// DecodeWithTTL decodes and verifies a cursor, also checking that it has not expired.
func (e *CursorEncoder) DecodeWithTTL(encoded string, ttl time.Duration) (*PaginationCursor, error) {
	cursor, err := e.Decode(encoded)
	if err != nil {
		return nil, err
	}

	if time.Since(cursor.CreatedAt) > ttl {
		return nil, fmt.Errorf("cursor expired: created %v ago, TTL is %v", time.Since(cursor.CreatedAt).Truncate(time.Second), ttl)
	}

	return cursor, nil
}

// DefaultPaginationConfig returns sensible defaults for pagination configuration.
func DefaultPaginationConfig() *PaginationConfig {
	return &PaginationConfig{
		DefaultPageSize: 20,
		MaxPageSize:     1000,
		CursorTTL:       24 * time.Hour,
		EnableCursor:    true,
		FallbackOffset:  true,
		Secret:          []byte("change-me-in-production"),
	}
}

// EnforcePageSize clamps the requested page size to the configured bounds.
// Returns DefaultPageSize if requested is 0, minimum of 1, and caps at MaxPageSize.
func EnforcePageSize(requested int, config *PaginationConfig) int {
	if requested == 0 {
		return config.DefaultPageSize
	}
	if requested < 0 {
		return 1
	}
	if config.MaxPageSize > 0 && requested > config.MaxPageSize {
		return config.MaxPageSize
	}
	return requested
}

// ParseCursorParams extracts cursor pagination parameters from query string values.
func ParseCursorParams(params url.Values) (*CursorParams, error) {
	cp := &CursorParams{
		After:  params.Get("_cursor"),
		Before: params.Get("_cursor:prev"),
		Sort:   params.Get("_sort"),
	}

	// Both cursors cannot be provided simultaneously
	if cp.After != "" && cp.Before != "" {
		return nil, fmt.Errorf("cannot specify both _cursor and _cursor:prev")
	}

	// Parse _count
	countStr := params.Get("_count")
	if countStr != "" {
		count, err := strconv.Atoi(countStr)
		if err != nil {
			return nil, fmt.Errorf("invalid _count value %q: %w", countStr, err)
		}
		if count < 0 {
			return nil, fmt.Errorf("_count must be non-negative, got %d", count)
		}
		cp.Count = count
	}

	return cp, nil
}

// ParseSortKeys parses a FHIR _sort parameter into SortKeys using the provided column map.
// Fields not found in the column map are silently ignored.
// The prefix "-" indicates descending; "+" or no prefix indicates ascending.
func ParseSortKeys(sortParam string, columnMap map[string]string) []SortKey {
	if sortParam == "" {
		return nil
	}

	parts := strings.Split(sortParam, ",")
	keys := make([]SortKey, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		ascending := true
		field := part

		if strings.HasPrefix(part, "-") {
			ascending = false
			field = part[1:]
		} else if strings.HasPrefix(part, "+") {
			field = part[1:]
		}

		col, ok := columnMap[field]
		if !ok {
			continue // skip unknown fields
		}

		keys = append(keys, SortKey{
			Field:     field,
			Column:    col,
			Ascending: ascending,
		})
	}

	return keys
}

// DefaultSortKeys returns the default sort order: lastUpdated DESC, id DESC.
func DefaultSortKeys() []SortKey {
	return []SortKey{
		{Field: "_lastUpdated", Column: "last_updated", Ascending: false},
		{Field: "_id", Column: "id", Ascending: false},
	}
}

// DefaultColumnMap maps common FHIR sort parameters to SQL column names.
func DefaultColumnMap() map[string]string {
	return map[string]string{
		"_lastUpdated": "last_updated",
		"_id":          "id",
		"date":         "effective_date",
		"name":         "family_name",
		"status":       "status",
		"identifier":   "identifier_value",
		"birthdate":    "birth_date",
		"gender":       "gender",
		"given":        "given_name",
		"family":       "family_name",
	}
}

// BuildKeysetWhereClause generates the WHERE clause for keyset pagination.
//
// For forward pagination with a single sort key (e.g., ORDER BY last_updated DESC, id DESC)
// and cursor at (t1, id1), it produces:
//
//	(last_updated, id) < ($startIdx, $startIdx+1)
//
// For ascending sort it uses > instead of <. For backward pagination the comparison
// is reversed. For multi-column sorts with mixed directions, a row-value comparison
// is decomposed into the standard keyset expansion.
//
// Returns empty string and nil args when cursor is nil (first page).
func BuildKeysetWhereClause(cursor *PaginationCursor, startIdx int) (string, []interface{}) {
	if cursor == nil || len(cursor.SortKeys) == 0 {
		return "", nil
	}

	// Determine whether we need to reverse the comparison operators
	// for backward pagination.
	reverse := cursor.Direction == CursorBackward

	// Build the multi-column keyset condition using the canonical expansion:
	// For ORDER BY a DESC, b ASC, id DESC with cursor (a0, b0, id0):
	// Forward:  (a < a0) OR (a = a0 AND b > b0) OR (a = a0 AND b = b0 AND id < id0)
	// Backward: (a > a0) OR (a = a0 AND b < b0) OR (a = a0 AND b = b0 AND id > id0)

	sortKeys := cursor.SortKeys
	var args []interface{}
	var orParts []string
	argIdx := startIdx

	// Collect all columns and values (sort keys + id tiebreaker)
	type colInfo struct {
		column    string
		value     interface{}
		ascending bool
		isNull    bool
	}

	cols := make([]colInfo, 0, len(sortKeys)+1)
	for _, sk := range sortKeys {
		val := cursor.Values[sk.Column]
		cols = append(cols, colInfo{
			column:    sk.Column,
			value:     val,
			ascending: sk.Ascending,
			isNull:    val == nil,
		})
	}
	// Add id as tiebreaker (always descending by convention)
	cols = append(cols, colInfo{
		column:    "id",
		value:     cursor.ID,
		ascending: false,
		isNull:    false,
	})

	// Build the expanded OR conditions
	for i := 0; i < len(cols); i++ {
		var andParts []string

		// All preceding columns must be equal
		for j := 0; j < i; j++ {
			if cols[j].isNull {
				andParts = append(andParts, fmt.Sprintf("%s IS NULL", cols[j].column))
			} else {
				andParts = append(andParts, fmt.Sprintf("%s = $%d", cols[j].column, argIdx))
				args = append(args, cols[j].value)
				argIdx++
			}
		}

		// The current column uses the inequality
		col := cols[i]
		if col.isNull {
			// If the cursor value is NULL:
			// - For ASC NULLS LAST (nulls are at end): in forward direction
			//   there's nothing after NULL, so this branch contributes nothing.
			// - For DESC NULLS LAST (nulls are at end): forward means we want
			//   rows before NULL, i.e., non-null rows: col IS NOT NULL
			if col.ascending {
				if reverse {
					andParts = append(andParts, fmt.Sprintf("%s IS NOT NULL", col.column))
				} else {
					// Nothing comes after NULL in ASC NULLS LAST, skip
					continue
				}
			} else {
				if reverse {
					// Nothing comes after NULL in DESC NULLS LAST for backward
					continue
				} else {
					andParts = append(andParts, fmt.Sprintf("%s IS NOT NULL", col.column))
				}
			}
		} else {
			// Determine comparison operator
			// ASC + forward  => >
			// ASC + backward => <
			// DESC + forward => <
			// DESC + backward => >
			var op string
			if col.ascending {
				if reverse {
					op = "<"
				} else {
					op = ">"
				}
			} else {
				if reverse {
					op = ">"
				} else {
					op = "<"
				}
			}
			andParts = append(andParts, fmt.Sprintf("%s %s $%d", col.column, op, argIdx))
			args = append(args, col.value)
			argIdx++
		}

		if len(andParts) > 0 {
			if len(andParts) == 1 {
				orParts = append(orParts, andParts[0])
			} else {
				orParts = append(orParts, "("+strings.Join(andParts, " AND ")+")")
			}
		}
	}

	if len(orParts) == 0 {
		return "", nil
	}

	clause := "(" + strings.Join(orParts, " OR ") + ")"
	return clause, args
}

// BuildKeysetOrderClause generates the ORDER BY clause for keyset pagination.
// Each sort key gets the appropriate direction and NULLS LAST treatment.
// An "id" tiebreaker is always appended if not already present.
func BuildKeysetOrderClause(sortKeys []SortKey) string {
	if len(sortKeys) == 0 {
		return ""
	}

	var parts []string
	hasID := false

	for _, sk := range sortKeys {
		dir := "ASC"
		if !sk.Ascending {
			dir = "DESC"
		}
		parts = append(parts, fmt.Sprintf("%s %s NULLS LAST", sk.Column, dir))
		if sk.Column == "id" {
			hasID = true
		}
	}

	// Always add id as tiebreaker
	if !hasID {
		parts = append(parts, "id DESC NULLS LAST")
	}

	return " ORDER BY " + strings.Join(parts, ", ")
}

// BuildCursorFromRow extracts cursor values from a resource row using the sort keys.
// Returns nil if the resource is nil.
func BuildCursorFromRow(resource map[string]interface{}, sortKeys []SortKey) *PaginationCursor {
	if resource == nil {
		return nil
	}

	cursor := &PaginationCursor{
		Values:   make(map[string]interface{}),
		SortKeys: sortKeys,
	}

	// Extract resource ID
	if id, ok := resource["id"].(string); ok {
		cursor.ID = id
	}

	// Extract sort key values from the resource
	for _, sk := range sortKeys {
		val := extractCursorFieldValue(resource, sk.Field, sk.Column)
		cursor.Values[sk.Column] = val
	}

	return cursor
}

// extractCursorFieldValue attempts to extract a value from a resource for a given FHIR field.
func extractCursorFieldValue(resource map[string]interface{}, field, column string) interface{} {
	// Try direct column name
	if v, ok := resource[column]; ok {
		return v
	}

	// Try FHIR field mapping
	switch field {
	case "_lastUpdated":
		if meta, ok := resource["meta"].(map[string]interface{}); ok {
			return meta["lastUpdated"]
		}
	case "_id":
		return resource["id"]
	default:
		// Try the field name directly
		if v, ok := resource[field]; ok {
			return v
		}
	}

	return nil
}

// BuildBundleLinks generates Bundle.link entries with cursor URLs for pagination.
// Returns a slice of maps with "relation" and "url" keys.
func BuildBundleLinks(baseURL string, page *CursorPage, encoder *CursorEncoder) []map[string]interface{} {
	links := []map[string]interface{}{
		{
			"relation": "self",
			"url":      fmt.Sprintf("%s?_count=%d", baseURL, page.PageSize),
		},
	}

	if page.HasNext && page.NextCursor != "" {
		links = append(links, map[string]interface{}{
			"relation": "next",
			"url":      fmt.Sprintf("%s?_count=%d&_cursor=%s", baseURL, page.PageSize, url.QueryEscape(page.NextCursor)),
		})
	}

	if page.HasPrev && page.PrevCursor != "" {
		links = append(links, map[string]interface{}{
			"relation": "previous",
			"url":      fmt.Sprintf("%s?_count=%d&_cursor:prev=%s", baseURL, page.PageSize, url.QueryEscape(page.PrevCursor)),
		})
	}

	return links
}

// ApplyCursorPagination modifies a SQL query to use cursor-based pagination.
// It appends a WHERE clause (if cursor is present), ORDER BY, and LIMIT.
// Returns the modified query and any args from the cursor.
func ApplyCursorPagination(query string, cursor *PaginationCursor, pageSize int, startIdx int) (string, []interface{}) {
	var args []interface{}
	var sortKeys []SortKey

	if cursor != nil && len(cursor.SortKeys) > 0 {
		sortKeys = cursor.SortKeys
	} else {
		sortKeys = DefaultSortKeys()
	}

	// Add keyset WHERE clause if cursor exists
	if cursor != nil {
		whereClause, cursorArgs := BuildKeysetWhereClause(cursor, startIdx)
		if whereClause != "" {
			if strings.Contains(strings.ToUpper(query), "WHERE") {
				query += " AND " + whereClause
			} else {
				query += " WHERE " + whereClause
			}
			args = append(args, cursorArgs...)
		}
	}

	// Add ORDER BY
	query += BuildKeysetOrderClause(sortKeys)

	// Add LIMIT (fetch one extra to detect hasNext)
	query += fmt.Sprintf(" LIMIT %d", pageSize+1)

	return query, args
}

// ValidateCursorConsistency checks that cursor sort keys match request sort keys.
// Both must have the same fields in the same order with the same directions.
func ValidateCursorConsistency(cursor *PaginationCursor, requestSort []SortKey) error {
	cursorKeys := cursor.SortKeys

	// Both empty is fine (both use defaults)
	if len(cursorKeys) == 0 && len(requestSort) == 0 {
		return nil
	}

	if len(cursorKeys) != len(requestSort) {
		return fmt.Errorf("cursor sort keys length (%d) does not match request sort keys length (%d)",
			len(cursorKeys), len(requestSort))
	}

	for i := range cursorKeys {
		if cursorKeys[i].Field != requestSort[i].Field {
			return fmt.Errorf("cursor sort key[%d] field %q does not match request field %q",
				i, cursorKeys[i].Field, requestSort[i].Field)
		}
		if cursorKeys[i].Ascending != requestSort[i].Ascending {
			return fmt.Errorf("cursor sort key[%d] direction mismatch for field %q",
				i, cursorKeys[i].Field)
		}
	}

	return nil
}

// EstimateTotal provides a fast approximate total count using PostgreSQL EXPLAIN.
// Returns an EXPLAIN query string that can be executed to get row estimates.
func EstimateTotal(query string) string {
	return fmt.Sprintf("EXPLAIN (FORMAT JSON) %s", query)
}

// PaginationMiddleware adds cursor pagination support to search endpoints.
// It parses cursor parameters, decodes cursors, and sets context values:
//   - "_cursorPageSize": enforced page size (int)
//   - "_paginationCursor": decoded *PaginationCursor (nil if first page)
//   - "_cursorEncoder": the *CursorEncoder for downstream use
func PaginationMiddleware(config *PaginationConfig) echo.MiddlewareFunc {
	encoder := NewCursorEncoder(config.Secret)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Only process GET and POST (search) requests
			if c.Request().Method != http.MethodGet {
				return next(c)
			}

			if !config.EnableCursor {
				return next(c)
			}

			// Parse cursor params
			params, err := ParseCursorParams(c.QueryParams())
			if err != nil {
				return c.JSON(http.StatusBadRequest, NewOperationOutcome(
					IssueSeverityError, IssueTypeInvalid, err.Error()))
			}

			// Enforce page size
			pageSize := EnforcePageSize(params.Count, config)
			c.Set("_cursorPageSize", pageSize)
			c.Set("_cursorEncoder", encoder)

			// Decode cursor if provided
			cursorToken := params.After
			direction := CursorForward
			if params.Before != "" {
				cursorToken = params.Before
				direction = CursorBackward
			}

			if cursorToken != "" {
				var cursor *PaginationCursor
				var decodeErr error

				if config.CursorTTL > 0 {
					cursor, decodeErr = encoder.DecodeWithTTL(cursorToken, config.CursorTTL)
				} else {
					cursor, decodeErr = encoder.Decode(cursorToken)
				}

				if decodeErr != nil {
					if config.FallbackOffset {
						// Fall back to offset pagination: do not set cursor context
						return next(c)
					}
					return c.JSON(http.StatusBadRequest, NewOperationOutcome(
						IssueSeverityError, IssueTypeInvalid,
						fmt.Sprintf("invalid pagination cursor: %v", decodeErr)))
				}

				cursor.Direction = direction
				c.Set("_paginationCursor", cursor)
			}

			return next(c)
		}
	}
}
