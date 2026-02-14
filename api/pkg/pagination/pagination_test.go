package pagination

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestFromContext_Defaults(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	p := FromContext(c)

	if p.Limit != DefaultLimit {
		t.Errorf("expected default limit %d, got %d", DefaultLimit, p.Limit)
	}
	if p.Offset != 0 {
		t.Errorf("expected default offset 0, got %d", p.Offset)
	}
}

func TestFromContext_CustomValues(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/?limit=50&offset=10", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	p := FromContext(c)

	if p.Limit != 50 {
		t.Errorf("expected limit 50, got %d", p.Limit)
	}
	if p.Offset != 10 {
		t.Errorf("expected offset 10, got %d", p.Offset)
	}
}

func TestFromContext_FHIRParams(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/?_count=25&_offset=5", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	p := FromContext(c)

	if p.Limit != 25 {
		t.Errorf("expected limit 25, got %d", p.Limit)
	}
	if p.Offset != 5 {
		t.Errorf("expected offset 5, got %d", p.Offset)
	}
}

func TestFromContext_MaxLimit(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/?limit=500", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	p := FromContext(c)

	if p.Limit != MaxLimit {
		t.Errorf("expected limit capped at %d, got %d", MaxLimit, p.Limit)
	}
}

func TestFromContext_NegativeOffset(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/?offset=-5", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	p := FromContext(c)

	if p.Offset != 0 {
		t.Errorf("expected offset 0 for negative input, got %d", p.Offset)
	}
}

func TestSQL(t *testing.T) {
	p := Params{Limit: 20, Offset: 40}
	expected := "LIMIT 20 OFFSET 40"
	if p.SQL() != expected {
		t.Errorf("expected %q, got %q", expected, p.SQL())
	}
}

func TestNewResponse(t *testing.T) {
	data := []string{"a", "b", "c"}
	r := NewResponse(data, 10, 3, 0)

	if r.Total != 10 {
		t.Errorf("expected total 10, got %d", r.Total)
	}
	if !r.HasMore {
		t.Error("expected has_more to be true when offset+limit < total")
	}

	r2 := NewResponse(data, 3, 3, 0)
	if r2.HasMore {
		t.Error("expected has_more to be false when offset+limit >= total")
	}
}

func TestParams_HasNext(t *testing.T) {
	tests := []struct {
		name   string
		params Params
		total  int
		want   bool
	}{
		{"more results", Params{Limit: 10, Offset: 0}, 25, true},
		{"exact end", Params{Limit: 10, Offset: 15}, 25, false},
		{"past end", Params{Limit: 10, Offset: 30}, 25, false},
		{"no results", Params{Limit: 10, Offset: 0}, 0, false},
		{"last partial page", Params{Limit: 10, Offset: 20}, 25, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.params.HasNext(tt.total); got != tt.want {
				t.Errorf("HasNext() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParams_HasPrevious(t *testing.T) {
	tests := []struct {
		name   string
		params Params
		want   bool
	}{
		{"first page", Params{Limit: 10, Offset: 0}, false},
		{"second page", Params{Limit: 10, Offset: 10}, true},
		{"middle", Params{Limit: 10, Offset: 25}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.params.HasPrevious(); got != tt.want {
				t.Errorf("HasPrevious() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParams_NextOffset(t *testing.T) {
	p := Params{Limit: 10, Offset: 5}
	if got := p.NextOffset(); got != 15 {
		t.Errorf("NextOffset() = %d, want 15", got)
	}
}

func TestParams_PreviousOffset(t *testing.T) {
	tests := []struct {
		name   string
		params Params
		want   int
	}{
		{"normal", Params{Limit: 10, Offset: 20}, 10},
		{"clamp to zero", Params{Limit: 10, Offset: 5}, 0},
		{"exact", Params{Limit: 10, Offset: 10}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.params.PreviousOffset(); got != tt.want {
				t.Errorf("PreviousOffset() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestParams_FHIRLinks_FirstPage(t *testing.T) {
	p := Params{Limit: 10, Offset: 0}
	links := p.FHIRLinks("/fhir/Patient", 25)

	linkMap := make(map[string]string)
	for _, l := range links {
		linkMap[l.Relation] = l.URL
	}

	if _, ok := linkMap["self"]; !ok {
		t.Error("expected 'self' link")
	}
	if _, ok := linkMap["next"]; !ok {
		t.Error("expected 'next' link")
	}
	if _, ok := linkMap["previous"]; ok {
		t.Error("did not expect 'previous' link on first page")
	}

	expectedSelf := "/fhir/Patient?_offset=0&_count=10"
	if linkMap["self"] != expectedSelf {
		t.Errorf("expected self %q, got %q", expectedSelf, linkMap["self"])
	}
	expectedNext := "/fhir/Patient?_offset=10&_count=10"
	if linkMap["next"] != expectedNext {
		t.Errorf("expected next %q, got %q", expectedNext, linkMap["next"])
	}
}

func TestParams_FHIRLinks_MiddlePage(t *testing.T) {
	p := Params{Limit: 10, Offset: 10}
	links := p.FHIRLinks("/fhir/Patient", 25)

	linkMap := make(map[string]string)
	for _, l := range links {
		linkMap[l.Relation] = l.URL
	}

	if _, ok := linkMap["self"]; !ok {
		t.Error("expected 'self' link")
	}
	if _, ok := linkMap["next"]; !ok {
		t.Error("expected 'next' link")
	}
	if _, ok := linkMap["previous"]; !ok {
		t.Error("expected 'previous' link")
	}

	expectedPrev := "/fhir/Patient?_offset=0&_count=10"
	if linkMap["previous"] != expectedPrev {
		t.Errorf("expected previous %q, got %q", expectedPrev, linkMap["previous"])
	}
}

func TestParams_FHIRLinks_LastPage(t *testing.T) {
	p := Params{Limit: 10, Offset: 20}
	links := p.FHIRLinks("/fhir/Patient", 25)

	linkMap := make(map[string]string)
	for _, l := range links {
		linkMap[l.Relation] = l.URL
	}

	if _, ok := linkMap["self"]; !ok {
		t.Error("expected 'self' link")
	}
	if _, ok := linkMap["next"]; ok {
		t.Error("did not expect 'next' link on last page")
	}
	if _, ok := linkMap["previous"]; !ok {
		t.Error("expected 'previous' link")
	}
}

func TestParams_FHIRLinks_NoResults(t *testing.T) {
	p := Params{Limit: 10, Offset: 0}
	links := p.FHIRLinks("/fhir/Patient", 0)

	if len(links) != 1 {
		t.Fatalf("expected 1 link (self only), got %d", len(links))
	}
	if links[0].Relation != "self" {
		t.Errorf("expected 'self', got %q", links[0].Relation)
	}
}

func TestFHIRLink_JSONFormat(t *testing.T) {
	link := FHIRLink{
		Relation: "next",
		URL:      "/fhir/Patient?_offset=20&_count=10",
	}
	if link.Relation != "next" {
		t.Errorf("expected relation 'next', got %q", link.Relation)
	}
	if link.URL != "/fhir/Patient?_offset=20&_count=10" {
		t.Errorf("unexpected URL: %q", link.URL)
	}
}
