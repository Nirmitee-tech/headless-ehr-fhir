package middleware

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

func Logger(logger zerolog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			req := c.Request()
			rid, _ := c.Get("request_id").(string)

			err := next(c)

			evt := logger.Info()
			if err != nil {
				evt = logger.Error().Err(err)
			}

			evt.
				Str("request_id", rid).
				Str("method", req.Method).
				Str("path", req.URL.Path).
				Int("status", c.Response().Status).
				Dur("latency", time.Since(start)).
				Str("remote_ip", c.RealIP()).
				Msg("request")

			return err
		}
	}
}
