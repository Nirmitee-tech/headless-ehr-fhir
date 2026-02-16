package fhir

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

// CodeSystemLookup can look up a code in a code system.
type CodeSystemLookup interface {
	LookupCode(system, code, version string) (*LookupResult, error)
}

// LookupResult represents the result of a CodeSystem $lookup operation.
type LookupResult struct {
	Name        string
	Version     string
	Display     string
	Abstract    bool
	Designation []LookupDesignation
	Property    []LookupProperty
}

// LookupDesignation represents an alternative display for a code.
type LookupDesignation struct {
	Language string
	Use      Coding
	Value    string
}

// LookupProperty represents a property of a code.
type LookupProperty struct {
	Code        string
	Value       interface{}
	Description string
}

// LookupHandler handles CodeSystem $lookup requests.
type LookupHandler struct {
	lookup CodeSystemLookup
}

// NewLookupHandler creates a new LookupHandler.
func NewLookupHandler(lookup CodeSystemLookup) *LookupHandler {
	return &LookupHandler{lookup: lookup}
}

// RegisterRoutes registers the $lookup endpoint.
func (h *LookupHandler) RegisterRoutes(fhirGroup *echo.Group) {
	fhirGroup.GET("/CodeSystem/$lookup", h.Lookup)
	fhirGroup.POST("/CodeSystem/$lookup", h.Lookup)
	fhirGroup.GET("/CodeSystem/:id/$lookup", h.LookupByID)
	fhirGroup.POST("/CodeSystem/:id/$lookup", h.LookupByID)
}

// Lookup handles GET/POST /fhir/CodeSystem/$lookup
func (h *LookupHandler) Lookup(c echo.Context) error {
	system := c.QueryParam("system")
	code := c.QueryParam("code")
	version := c.QueryParam("version")

	if code == "" {
		return c.JSON(http.StatusBadRequest, ErrorOutcome("code parameter is required for CodeSystem $lookup"))
	}

	return h.doLookup(c, system, code, version)
}

// LookupByID handles GET/POST /fhir/CodeSystem/:id/$lookup
func (h *LookupHandler) LookupByID(c echo.Context) error {
	system := c.Param("id")
	code := c.QueryParam("code")
	version := c.QueryParam("version")

	if code == "" {
		return c.JSON(http.StatusBadRequest, ErrorOutcome("code parameter is required for CodeSystem $lookup"))
	}

	return h.doLookup(c, system, code, version)
}

func (h *LookupHandler) doLookup(c echo.Context, system, code, version string) error {
	result, err := h.lookup.LookupCode(system, code, version)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return c.JSON(http.StatusNotFound, ErrorOutcome("code not found: "+code))
		}
		return c.JSON(http.StatusInternalServerError, ErrorOutcome(err.Error()))
	}

	return c.JSON(http.StatusOK, h.toFHIR(result))
}

func (h *LookupHandler) toFHIR(r *LookupResult) map[string]interface{} {
	result := map[string]interface{}{
		"resourceType": "Parameters",
	}

	params := []interface{}{}

	if r.Name != "" {
		params = append(params, map[string]interface{}{
			"name":        "name",
			"valueString": r.Name,
		})
	}
	if r.Version != "" {
		params = append(params, map[string]interface{}{
			"name":        "version",
			"valueString": r.Version,
		})
	}
	if r.Display != "" {
		params = append(params, map[string]interface{}{
			"name":        "display",
			"valueString": r.Display,
		})
	}
	params = append(params, map[string]interface{}{
		"name":         "abstract",
		"valueBoolean": r.Abstract,
	})

	for _, d := range r.Designation {
		parts := []interface{}{}
		if d.Language != "" {
			parts = append(parts, map[string]interface{}{
				"name":      "language",
				"valueCode": d.Language,
			})
		}
		if d.Use.Code != "" {
			parts = append(parts, map[string]interface{}{
				"name": "use",
				"valueCoding": map[string]interface{}{
					"system":  d.Use.System,
					"code":    d.Use.Code,
					"display": d.Use.Display,
				},
			})
		}
		if d.Value != "" {
			parts = append(parts, map[string]interface{}{
				"name":        "value",
				"valueString": d.Value,
			})
		}
		params = append(params, map[string]interface{}{
			"name": "designation",
			"part": parts,
		})
	}

	for _, p := range r.Property {
		parts := []interface{}{
			map[string]interface{}{
				"name":      "code",
				"valueCode": p.Code,
			},
		}
		switch v := p.Value.(type) {
		case string:
			parts = append(parts, map[string]interface{}{
				"name":        "value",
				"valueString": v,
			})
		case bool:
			parts = append(parts, map[string]interface{}{
				"name":         "value",
				"valueBoolean": v,
			})
		case int:
			parts = append(parts, map[string]interface{}{
				"name":         "value",
				"valueInteger": v,
			})
		}
		if p.Description != "" {
			parts = append(parts, map[string]interface{}{
				"name":        "description",
				"valueString": p.Description,
			})
		}
		params = append(params, map[string]interface{}{
			"name": "property",
			"part": parts,
		})
	}

	result["parameter"] = params
	return result
}
