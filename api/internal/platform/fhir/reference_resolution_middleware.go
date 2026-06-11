package fhir

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"
)

// ReferenceResolutionMiddleware rewrites internal-UUID references in FHIR
// responses to external fhir_id references (see reference_resolver.go for the
// rationale). It mirrors ProjectionMiddleware: it records the response body,
// and if the body is a FHIR resource or searchset Bundle, resolves references
// in each resource before writing.
//
// newResolver constructs a fresh ReferenceResolver per request, so the
// resolver's lookup cache lives only as long as the response being built. The
// lookup itself (DB-backed) is supplied at wiring time and uses the request's
// tenant-scoped connection.
func ReferenceResolutionMiddleware(newResolver func() *ReferenceResolver) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			rec := &referenceResponseRecorder{
				ResponseWriter: c.Response().Writer,
				body:           &bytes.Buffer{},
			}
			c.Response().Writer = rec

			if err := next(c); err != nil {
				return err
			}

			// Only attempt JSON resources; pass through everything else verbatim
			// (errors, non-JSON, empty bodies) without altering status codes.
			var resource map[string]interface{}
			if rec.body.Len() == 0 || json.Unmarshal(rec.body.Bytes(), &resource) != nil {
				return writeRecorded(rec)
			}

			resolver := newResolver()
			ctx := c.Request().Context()

			switch resource["resourceType"] {
			case "Bundle":
				if entries, ok := resource["entry"].([]interface{}); ok {
					for _, entry := range entries {
						entryMap, ok := entry.(map[string]interface{})
						if !ok {
							continue
						}
						if res, ok := entryMap["resource"].(map[string]interface{}); ok {
							resolver.Resolve(ctx, res)
						}
					}
				}
			default:
				if resource["resourceType"] != nil {
					resolver.Resolve(ctx, resource)
				} else {
					// Not a FHIR resource (no resourceType) — leave untouched.
					return writeRecorded(rec)
				}
			}

			out, err := json.Marshal(resource)
			if err != nil {
				return writeRecorded(rec)
			}
			rec.flushHeader()
			_, writeErr := rec.ResponseWriter.Write(out)
			return writeErr
		}
	}
}

// writeRecorded flushes the originally recorded body unchanged.
func writeRecorded(rec *referenceResponseRecorder) error {
	rec.flushHeader()
	_, err := rec.ResponseWriter.Write(rec.body.Bytes())
	return err
}

// referenceResponseRecorder buffers the response body for post-processing,
// deferring the status code so it is written exactly once alongside the
// (possibly rewritten) body.
type referenceResponseRecorder struct {
	http.ResponseWriter
	body       *bytes.Buffer
	statusCode int
	flushed    bool
}

func (r *referenceResponseRecorder) Write(b []byte) (int, error) {
	return r.body.Write(b)
}

func (r *referenceResponseRecorder) WriteHeader(code int) {
	r.statusCode = code
}

// flushHeader writes the recorded status (default 200) to the real writer
// exactly once. Echo has already marked its own Response committed, so we
// write the status straight to the underlying writer.
func (r *referenceResponseRecorder) flushHeader() {
	if r.flushed {
		return
	}
	r.flushed = true
	code := r.statusCode
	if code == 0 {
		code = http.StatusOK
	}
	r.ResponseWriter.WriteHeader(code)
}
