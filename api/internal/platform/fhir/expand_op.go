package fhir

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

// ValueSetExpander can expand a ValueSet by URL or ID.
type ValueSetExpander interface {
	ExpandValueSet(url string, filter string, offset, count int) (*ExpandedValueSet, error)
}

// ExpandedValueSet represents the result of a $expand operation.
type ExpandedValueSet struct {
	URL         string
	Version     string
	Name        string
	Title       string
	Status      string
	Total       int
	Offset      int
	Contains    []ValueSetContains
}

// ValueSetContains represents a concept within an expanded ValueSet.
type ValueSetContains struct {
	System    string
	Version   string
	Code      string
	Display   string
	Abstract  bool
	Inactive  bool
	Contains  []ValueSetContains
}

// ExpandHandler handles ValueSet $expand requests.
type ExpandHandler struct {
	expander ValueSetExpander
}

// NewExpandHandler creates a new ExpandHandler.
func NewExpandHandler(expander ValueSetExpander) *ExpandHandler {
	return &ExpandHandler{expander: expander}
}

// RegisterRoutes registers the $expand endpoint.
func (h *ExpandHandler) RegisterRoutes(fhirGroup *echo.Group) {
	fhirGroup.GET("/ValueSet/$expand", h.Expand)
	fhirGroup.POST("/ValueSet/$expand", h.Expand)
	fhirGroup.GET("/ValueSet/:id/$expand", h.ExpandByID)
	fhirGroup.POST("/ValueSet/:id/$expand", h.ExpandByID)
}

// Expand handles GET/POST /fhir/ValueSet/$expand
func (h *ExpandHandler) Expand(c echo.Context) error {
	url := c.QueryParam("url")
	if url == "" {
		return c.JSON(http.StatusBadRequest, ErrorOutcome("url parameter is required for ValueSet $expand"))
	}
	return h.doExpand(c, url)
}

// ExpandByID handles GET/POST /fhir/ValueSet/:id/$expand
func (h *ExpandHandler) ExpandByID(c echo.Context) error {
	id := c.Param("id")
	return h.doExpand(c, id)
}

func (h *ExpandHandler) doExpand(c echo.Context, urlOrID string) error {
	filter := c.QueryParam("filter")
	offset := intParam(c, "offset", 0)
	count := intParam(c, "count", 100)

	expanded, err := h.expander.ExpandValueSet(urlOrID, filter, offset, count)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return c.JSON(http.StatusNotFound, NotFoundOutcome("ValueSet", urlOrID))
		}
		return c.JSON(http.StatusInternalServerError, ErrorOutcome(err.Error()))
	}

	return c.JSON(http.StatusOK, h.toFHIR(expanded))
}

func (h *ExpandHandler) toFHIR(vs *ExpandedValueSet) map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "ValueSet",
		"status":       vs.Status,
	}
	if vs.URL != "" {
		result["url"] = vs.URL
	}
	if vs.Version != "" {
		result["version"] = vs.Version
	}
	if vs.Name != "" {
		result["name"] = vs.Name
	}
	if vs.Title != "" {
		result["title"] = vs.Title
	}

	expansion := map[string]interface{}{
		"total":  vs.Total,
		"offset": vs.Offset,
	}

	if len(vs.Contains) > 0 {
		expansion["contains"] = h.containsToFHIR(vs.Contains)
	}

	result["expansion"] = expansion
	return result
}

func (h *ExpandHandler) containsToFHIR(items []ValueSetContains) []interface{} {
	result := make([]interface{}, 0, len(items))
	for _, item := range items {
		entry := map[string]interface{}{}
		if item.System != "" {
			entry["system"] = item.System
		}
		if item.Version != "" {
			entry["version"] = item.Version
		}
		if item.Code != "" {
			entry["code"] = item.Code
		}
		if item.Display != "" {
			entry["display"] = item.Display
		}
		if item.Abstract {
			entry["abstract"] = true
		}
		if item.Inactive {
			entry["inactive"] = true
		}
		if len(item.Contains) > 0 {
			entry["contains"] = h.containsToFHIR(item.Contains)
		}
		result = append(result, entry)
	}
	return result
}

func intParam(c echo.Context, name string, defaultValue int) int {
	v := c.QueryParam(name)
	if v == "" {
		return defaultValue
	}
	n := 0
	for _, ch := range v {
		if ch >= '0' && ch <= '9' {
			n = n*10 + int(ch-'0')
		} else {
			return defaultValue
		}
	}
	return n
}
