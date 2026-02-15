package fhir

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"
)

// IncludeMiddleware creates Echo middleware that processes _include and _revinclude
// parameters on FHIR search responses. It intercepts search bundle responses and
// resolves included resources by appending them to the bundle entries.
//
// The middleware is a no-op when _include/_revinclude query params are absent,
// adding zero overhead to normal search requests.
func IncludeMiddleware(registry *IncludeRegistry) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			includeParams := c.QueryParams()["_include"]
			revIncludeParams := c.QueryParams()["_revinclude"]

			// Fast path: no include params, pass through with zero overhead.
			if len(includeParams) == 0 && len(revIncludeParams) == 0 {
				return next(c)
			}

			// Capture the response body by wrapping the response writer.
			origWriter := c.Response().Writer
			rec := &responseRecorder{
				ResponseWriter: origWriter,
				body:           &bytes.Buffer{},
				statusCode:     http.StatusOK,
			}
			c.Response().Writer = rec

			// Execute the handler chain.
			if err := next(c); err != nil {
				// Restore the original writer before returning the error so
				// the Echo error handler can write directly.
				c.Response().Writer = origWriter
				return err
			}

			// Only process JSON search bundles with 2xx status codes.
			if rec.statusCode < 200 || rec.statusCode >= 300 {
				return flushOriginal(origWriter, rec)
			}

			// Try to parse the response as a FHIR Bundle.
			var bundle Bundle
			if err := json.Unmarshal(rec.body.Bytes(), &bundle); err != nil {
				return flushOriginal(origWriter, rec)
			}

			// Only process searchset bundles.
			if bundle.Type != "searchset" {
				return flushOriginal(origWriter, rec)
			}

			// Extract matched resources from the bundle entries for include resolution.
			var matchedResources []interface{}
			for _, entry := range bundle.Entry {
				if entry.Resource != nil {
					var res map[string]interface{}
					if err := json.Unmarshal(entry.Resource, &res); err == nil {
						matchedResources = append(matchedResources, res)
					}
				}
			}

			if len(matchedResources) == 0 {
				return flushOriginal(origWriter, rec)
			}

			ctx := c.Request().Context()

			// Resolve _include params.
			if len(includeParams) > 0 {
				included, err := registry.ResolveIncludes(ctx, matchedResources, includeParams)
				if err == nil && len(included) > 0 {
					bundle.Entry = append(bundle.Entry, included...)
				}
				// On error, we still return the original results (resilient behavior).
			}

			// Resolve _revinclude params. _revinclude uses the same mechanism
			// but the semantics differ: the include params reference resources that
			// point back to the matched resources. We pass them through the same
			// ResolveIncludes path because the registry already has the reverse
			// reference definitions registered.
			if len(revIncludeParams) > 0 {
				revIncluded, err := registry.ResolveIncludes(ctx, matchedResources, revIncludeParams)
				if err == nil && len(revIncluded) > 0 {
					bundle.Entry = append(bundle.Entry, revIncluded...)
				}
			}

			// Re-serialize and write the augmented bundle.
			augmented, err := json.Marshal(bundle)
			if err != nil {
				return flushOriginal(origWriter, rec)
			}

			origWriter.Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
			origWriter.WriteHeader(rec.statusCode)
			_, writeErr := origWriter.Write(augmented)
			return writeErr
		}
	}
}

// responseRecorder captures the response body and status code written by
// downstream handlers so the middleware can inspect and modify the output.
type responseRecorder struct {
	http.ResponseWriter
	body       *bytes.Buffer
	statusCode int
	wroteHead  bool
}

func (r *responseRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.wroteHead = true
	// Do NOT forward to the underlying writer yet; we buffer everything.
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	if !r.wroteHead {
		r.statusCode = http.StatusOK
		r.wroteHead = true
	}
	return r.body.Write(b)
}

// flushOriginal writes the buffered response directly to the original
// http.ResponseWriter, bypassing Echo's Response wrapper to avoid issues
// with already-committed state.
func flushOriginal(w http.ResponseWriter, rec *responseRecorder) error {
	// Copy headers that the handler may have set via the recorder.
	for k, vals := range rec.Header() {
		for _, v := range vals {
			w.Header().Set(k, v)
		}
	}
	w.WriteHeader(rec.statusCode)
	_, err := w.Write(rec.body.Bytes())
	return err
}
