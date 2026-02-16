package fhir

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

// SearchMiddleware applies FHIR search parameter processing to responses.
// It handles _elements, _summary, _count, _total, and _sort post-processing.
// This middleware wraps search responses to apply projection and pagination.
func SearchMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Check if this is a search request
			if c.Request().Method != http.MethodGet && c.Request().Method != http.MethodPost {
				return next(c)
			}

			elements := c.QueryParam("_elements")
			summary := c.QueryParam("_summary")
			total := c.QueryParam("_total")

			// If no projection params, just pass through
			if elements == "" && summary == "" && total == "" {
				return next(c)
			}

			// Use a response recorder to capture the output
			origWriter := c.Response().Writer
			rec := &searchResponseRecorder{
				ResponseWriter: origWriter,
				body:           &bytes.Buffer{},
				statusCode:     http.StatusOK,
			}
			c.Response().Writer = rec

			if err := next(c); err != nil {
				c.Response().Writer = origWriter
				return err
			}

			// Parse the response body as a Bundle
			if rec.statusCode < 200 || rec.statusCode >= 300 {
				// Error response -- pass through
				return flushSearchRecorder(origWriter, rec)
			}

			var bundle Bundle
			if err := json.Unmarshal(rec.body.Bytes(), &bundle); err != nil {
				return flushSearchRecorder(origWriter, rec)
			}

			if bundle.ResourceType != "Bundle" {
				return flushSearchRecorder(origWriter, rec)
			}

			// Apply projection to bundle entries
			ApplyProjectionToBundle(&bundle, elements, summary)

			// Handle _total=none (remove total from bundle)
			if total == "none" {
				bundle.Total = nil
			}

			result, err := json.Marshal(bundle)
			if err != nil {
				return flushSearchRecorder(origWriter, rec)
			}

			origWriter.Header().Set(echo.HeaderContentType, "application/fhir+json; charset=utf-8")
			origWriter.WriteHeader(rec.statusCode)
			_, writeErr := origWriter.Write(result)
			return writeErr
		}
	}
}

// ParseCount extracts the _count parameter (defaults to limit if not provided).
func ParseCount(c echo.Context, defaultCount int) int {
	countStr := c.QueryParam("_count")
	if countStr == "" {
		return defaultCount
	}
	count, err := strconv.Atoi(countStr)
	if err != nil || count < 0 {
		return defaultCount
	}
	if count == 0 {
		return 0 // _count=0 means return count only
	}
	return count
}

// ParseOffset extracts the _offset parameter (defaults to 0).
func ParseOffset(c echo.Context) int {
	offsetStr := c.QueryParam("_offset")
	if offsetStr == "" {
		return 0
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		return 0
	}
	return offset
}

// searchResponseRecorder captures HTTP response data for post-processing
// in the search middleware pipeline.
type searchResponseRecorder struct {
	http.ResponseWriter
	body       *bytes.Buffer
	statusCode int
	wroteHead  bool
}

func (r *searchResponseRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.wroteHead = true
	// Do NOT forward to the underlying writer yet; we buffer everything.
}

func (r *searchResponseRecorder) Write(b []byte) (int, error) {
	if !r.wroteHead {
		r.statusCode = http.StatusOK
		r.wroteHead = true
	}
	return r.body.Write(b)
}

// flushSearchRecorder writes the buffered response directly to the original
// http.ResponseWriter, bypassing Echo's Response wrapper.
func flushSearchRecorder(w http.ResponseWriter, rec *searchResponseRecorder) error {
	for k, vals := range rec.Header() {
		for _, v := range vals {
			w.Header().Set(k, v)
		}
	}
	w.WriteHeader(rec.statusCode)
	_, err := w.Write(rec.body.Bytes())
	return err
}
