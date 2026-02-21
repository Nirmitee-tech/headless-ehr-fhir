package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

// RequestTimeout returns middleware that sets a context deadline on each
// incoming request. If the deadline is exceeded before the handler completes,
// the request context is cancelled and a 504 Gateway Timeout response with a
// FHIR OperationOutcome body is returned.
//
// WebSocket and SSE connections (paths starting with /ws/) are excluded because
// they are long-lived by design. Individual handlers that need more time (e.g.
// $export) can derive a new context with a longer deadline from the request
// context.
func RequestTimeout(timeout time.Duration) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Skip timeout for WebSocket / SSE paths
			if strings.HasPrefix(c.Request().URL.Path, "/ws/") {
				return next(c)
			}

			ctx, cancel := context.WithTimeout(c.Request().Context(), timeout)
			defer cancel()

			c.SetRequest(c.Request().WithContext(ctx))

			// Run handler in a goroutine so we can select on the context.
			done := make(chan error, 1)
			go func() {
				done <- next(c)
			}()

			select {
			case err := <-done:
				return err
			case <-ctx.Done():
				// If the context was cancelled due to timeout, return 504.
				if ctx.Err() == context.DeadlineExceeded {
					return gatewayTimeoutError(c)
				}
				// For other cancellation reasons (e.g. client disconnect),
				// just return the context error.
				return ctx.Err()
			}
		}
	}
}

// gatewayTimeoutError returns a 504 response with a FHIR OperationOutcome.
func gatewayTimeoutError(c echo.Context) error {
	outcome := map[string]interface{}{
		"resourceType": "OperationOutcome",
		"issue": []map[string]interface{}{
			{
				"severity":    "error",
				"code":        "timeout",
				"diagnostics": "Request processing exceeded the allowed time limit",
			},
		},
	}
	// Attempt to write the timeout response. If the response was already
	// committed (partial write), this will be a no-op.
	if !c.Response().Committed {
		return c.JSON(http.StatusGatewayTimeout, outcome)
	}
	return nil
}
