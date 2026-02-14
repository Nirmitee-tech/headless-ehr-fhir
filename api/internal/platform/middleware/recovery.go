package middleware

import (
	"fmt"
	"net/http"
	"runtime"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

func Recovery(logger zerolog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			defer func() {
				if r := recover(); r != nil {
					var stack [4096]byte
					n := runtime.Stack(stack[:], false)

					logger.Error().
						Str("request_id", fmt.Sprintf("%v", c.Get("request_id"))).
						Str("panic", fmt.Sprintf("%v", r)).
						Str("stack", string(stack[:n])).
						Msg("panic recovered")

					err = echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
				}
			}()
			return next(c)
		}
	}
}
