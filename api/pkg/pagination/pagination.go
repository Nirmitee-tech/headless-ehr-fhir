package pagination

import (
	"fmt"
	"strconv"

	"github.com/labstack/echo/v4"
)

const (
	DefaultLimit = 20
	MaxLimit     = 100
)

// Params holds pagination parameters extracted from a request.
type Params struct {
	Limit  int
	Offset int
}

// FromContext extracts pagination parameters from the echo context.
func FromContext(c echo.Context) Params {
	limit, _ := strconv.Atoi(c.QueryParam("_count"))
	if limit <= 0 {
		limit, _ = strconv.Atoi(c.QueryParam("limit"))
	}
	if limit <= 0 {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}

	offset, _ := strconv.Atoi(c.QueryParam("_offset"))
	if offset <= 0 {
		offset, _ = strconv.Atoi(c.QueryParam("offset"))
	}
	if offset < 0 {
		offset = 0
	}

	return Params{Limit: limit, Offset: offset}
}

// Response wraps a paginated API response.
type Response struct {
	Data       interface{} `json:"data"`
	Total      int         `json:"total"`
	Limit      int         `json:"limit"`
	Offset     int         `json:"offset"`
	HasMore    bool        `json:"has_more"`
}

func NewResponse(data interface{}, total, limit, offset int) *Response {
	return &Response{
		Data:    data,
		Total:   total,
		Limit:   limit,
		Offset:  offset,
		HasMore: offset+limit < total,
	}
}

// SQL returns the LIMIT and OFFSET clause for SQL queries.
func (p Params) SQL() string {
	return fmt.Sprintf("LIMIT %d OFFSET %d", p.Limit, p.Offset)
}

// HasNext returns true if there are more results after the current page.
func (p Params) HasNext(total int) bool {
	return p.Offset+p.Limit < total
}

// HasPrevious returns true if there are results before the current page.
func (p Params) HasPrevious() bool {
	return p.Offset > 0
}

// NextOffset returns the offset for the next page.
func (p Params) NextOffset() int {
	return p.Offset + p.Limit
}

// PreviousOffset returns the offset for the previous page.
// Returns 0 if the result would be negative.
func (p Params) PreviousOffset() int {
	prev := p.Offset - p.Limit
	if prev < 0 {
		return 0
	}
	return prev
}

// FHIRLinks generates FHIR Bundle pagination links for a search result.
// basePath should be the request path (e.g., "/fhir/Patient").
// Additional query params (filters) can be passed as extraParams in "key=value" format.
func (p Params) FHIRLinks(basePath string, total int) []FHIRLink {
	links := []FHIRLink{
		{
			Relation: "self",
			URL:      fmt.Sprintf("%s?_offset=%d&_count=%d", basePath, p.Offset, p.Limit),
		},
	}

	if p.HasNext(total) {
		links = append(links, FHIRLink{
			Relation: "next",
			URL:      fmt.Sprintf("%s?_offset=%d&_count=%d", basePath, p.NextOffset(), p.Limit),
		})
	}

	if p.HasPrevious() {
		links = append(links, FHIRLink{
			Relation: "previous",
			URL:      fmt.Sprintf("%s?_offset=%d&_count=%d", basePath, p.PreviousOffset(), p.Limit),
		})
	}

	return links
}

// FHIRLink represents a single FHIR Bundle link entry.
type FHIRLink struct {
	Relation string `json:"relation"`
	URL      string `json:"url"`
}
