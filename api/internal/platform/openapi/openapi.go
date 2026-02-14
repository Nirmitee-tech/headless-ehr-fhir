package openapi

import (
	"net/http"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/labstack/echo/v4"
)

// Generator builds an OpenAPI 3.0 spec from the CapabilityBuilder.
type Generator struct {
	capBuilder *fhir.CapabilityBuilder
	version    string
	baseURL    string
}

// NewGenerator creates a new OpenAPI spec generator.
func NewGenerator(capBuilder *fhir.CapabilityBuilder, version, baseURL string) *Generator {
	return &Generator{capBuilder: capBuilder, version: version, baseURL: baseURL}
}

// GenerateSpec produces the OpenAPI 3.0 spec as a map.
func (g *Generator) GenerateSpec() map[string]interface{} {
	cap := g.capBuilder.Build()

	paths := make(map[string]interface{})

	// Extract resources from capability statement
	restArray, _ := cap["rest"].([]map[string]interface{})
	if len(restArray) > 0 {
		resources, _ := restArray[0]["resource"].([]map[string]interface{})
		for _, res := range resources {
			resType, _ := res["type"].(string)
			if resType == "" {
				continue
			}

			// Add search path
			searchPath := "/fhir/" + resType
			paths[searchPath] = map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "Search " + resType,
					"operationId": "search" + resType,
					"tags":        []string{resType},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Search results Bundle",
						},
					},
				},
				"post": map[string]interface{}{
					"summary":     "Create " + resType,
					"operationId": "create" + resType,
					"tags":        []string{resType},
					"responses": map[string]interface{}{
						"201": map[string]interface{}{
							"description": "Created",
						},
					},
				},
			}

			// Add read path
			readPath := "/fhir/" + resType + "/{id}"
			paths[readPath] = map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "Read " + resType,
					"operationId": "read" + resType,
					"tags":        []string{resType},
					"parameters": []map[string]interface{}{
						{"name": "id", "in": "path", "required": true, "schema": map[string]string{"type": "string"}},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{"description": "Success"},
						"404": map[string]interface{}{"description": "Not Found"},
					},
				},
				"put": map[string]interface{}{
					"summary":     "Update " + resType,
					"operationId": "update" + resType,
					"tags":        []string{resType},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{"description": "Updated"},
					},
				},
				"delete": map[string]interface{}{
					"summary":     "Delete " + resType,
					"operationId": "delete" + resType,
					"tags":        []string{resType},
					"responses": map[string]interface{}{
						"204": map[string]interface{}{"description": "Deleted"},
					},
				},
			}
		}
	}

	spec := map[string]interface{}{
		"openapi": "3.0.3",
		"info": map[string]interface{}{
			"title":       "Headless EHR FHIR R4 API",
			"version":     g.version,
			"description": "FHIR R4 compliant EHR API",
		},
		"servers": []map[string]string{
			{"url": g.baseURL},
		},
		"paths": paths,
	}

	return spec
}

// RegisterRoutes registers the OpenAPI endpoints.
func (g *Generator) RegisterRoutes(apiGroup *echo.Group) {
	apiGroup.GET("/openapi.json", func(c echo.Context) error {
		return c.JSON(http.StatusOK, g.GenerateSpec())
	})
	apiGroup.GET("/docs", func(c echo.Context) error {
		return c.Redirect(http.StatusFound, "https://petstore.swagger.io/?url="+g.baseURL+"/api/openapi.json")
	})
}
