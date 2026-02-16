package fhir

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/labstack/echo/v4"
)

// Cursor represents a keyset pagination cursor encoding a sort value and resource ID.
// The pair (sort_value, id) uniquely identifies a position in a sorted result set,
// enabling efficient cursor-based (keyset) pagination without offset skipping.
type Cursor struct {
	Value string `json:"v"`
	ID    string `json:"id"`
}

// EncodeCursor encodes a sort value and resource ID into an opaque base64 cursor token.
func EncodeCursor(sortValue string, id string) string {
	c := Cursor{Value: sortValue, ID: id}
	data, _ := json.Marshal(c)
	return base64.RawURLEncoding.EncodeToString(data)
}

// DecodeCursor decodes an opaque cursor token back into a Cursor struct.
// Returns an error if the token is not valid base64 or does not contain valid JSON.
func DecodeCursor(token string) (*Cursor, error) {
	data, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return nil, fmt.Errorf("invalid cursor token: %w", err)
	}

	var c Cursor
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("invalid cursor payload: %w", err)
	}

	return &c, nil
}

// CursorBundleParams holds parameters for building a searchset Bundle with cursor-based pagination.
type CursorBundleParams struct {
	BaseURL    string
	QueryStr   string
	Count      int
	Total      int
	HasMore    bool
	NextCursor string // opaque cursor token for next page
}

// NewSearchBundleWithCursor creates a searchset Bundle using cursor-based pagination.
// Pagination links use _pageToken instead of _offset for keyset pagination.
func NewSearchBundleWithCursor(resources []interface{}, params CursorBundleParams) *Bundle {
	now := time.Now().UTC()
	entries := make([]BundleEntry, len(resources))
	for i, r := range resources {
		raw, _ := json.Marshal(r)
		fullURL := extractFullURL(r, params.BaseURL)
		entries[i] = BundleEntry{
			FullURL:  fullURL,
			Resource: raw,
			Search: &BundleSearch{
				Mode: "match",
			},
		}
	}

	links := buildCursorPaginationLinks(params)

	return &Bundle{
		ResourceType: "Bundle",
		Type:         "searchset",
		Total:        &params.Total,
		Timestamp:    &now,
		Link:         links,
		Entry:        entries,
	}
}

// buildCursorPaginationLinks creates self and next links for cursor-based pagination.
func buildCursorPaginationLinks(params CursorBundleParams) []BundleLink {
	links := []BundleLink{
		{
			Relation: "self",
			URL:      fmt.Sprintf("%s?%s_count=%d", params.BaseURL, conditionalAmpersand(params.QueryStr), params.Count),
		},
	}

	if params.HasMore && params.NextCursor != "" {
		links = append(links, BundleLink{
			Relation: "next",
			URL:      fmt.Sprintf("%s?%s_count=%d&_pageToken=%s", params.BaseURL, conditionalAmpersand(params.QueryStr), params.Count, params.NextCursor),
		})
	}

	return links
}

// ParsePageToken extracts the _pageToken query parameter from the request.
// Returns an empty string if the parameter is not present.
func ParsePageToken(c echo.Context) string {
	return c.QueryParam("_pageToken")
}
