package fhir

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"
)

// ProjectionMiddleware applies _summary and _elements filtering to FHIR responses.
// It intercepts the response body, parses it as JSON, applies the projection,
// and writes the modified response. For Bundle resources, projection is applied
// to each entry's resource.
func ProjectionMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			summary := c.QueryParam("_summary")
			elements := c.QueryParam("_elements")

			if summary == "" && elements == "" {
				return next(c)
			}

			// Create a response recorder
			rec := &projectionResponseRecorder{
				ResponseWriter: c.Response().Writer,
				body:           &bytes.Buffer{},
			}
			c.Response().Writer = rec

			if err := next(c); err != nil {
				return err
			}

			// Try to apply projection to the response
			var resource map[string]interface{}
			if err := json.Unmarshal(rec.body.Bytes(), &resource); err != nil {
				// Not JSON or not a map, write original
				c.Response().Writer = rec.ResponseWriter
				_, writeErr := c.Response().Writer.Write(rec.body.Bytes())
				return writeErr
			}

			// Check if it's a Bundle (apply to entries) or a single resource
			if resource["resourceType"] == "Bundle" {
				// Apply to each entry
				if entries, ok := resource["entry"].([]interface{}); ok {
					for i, entry := range entries {
						if entryMap, ok := entry.(map[string]interface{}); ok {
							if res, ok := entryMap["resource"].(map[string]interface{}); ok {
								entryMap["resource"] = ApplyProjection(res, elements, summary)
								entries[i] = entryMap
							}
						}
					}
					resource["entry"] = entries
				}
			} else {
				resource = ApplyProjection(resource, elements, summary)
			}

			result, err := json.Marshal(resource)
			if err != nil {
				// If marshal fails, write original
				c.Response().Writer = rec.ResponseWriter
				_, writeErr := c.Response().Writer.Write(rec.body.Bytes())
				return writeErr
			}

			c.Response().Writer = rec.ResponseWriter
			c.Response().Header().Set("Content-Type", "application/fhir+json")
			_, writeErr := c.Response().Writer.Write(result)
			return writeErr
		}
	}
}

// projectionResponseRecorder captures the response body for post-processing.
type projectionResponseRecorder struct {
	http.ResponseWriter
	body       *bytes.Buffer
	statusCode int
}

func (r *projectionResponseRecorder) Write(b []byte) (int, error) {
	return r.body.Write(b)
}

func (r *projectionResponseRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}
