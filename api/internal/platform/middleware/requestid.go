package middleware

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

const RequestIDHeader = "X-Request-ID"

func RequestID() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			rid := c.Request().Header.Get(RequestIDHeader)
			if rid == "" {
				rid = uuid.New().String()
			}
			c.Set("request_id", rid)
			c.Response().Header().Set(RequestIDHeader, rid)
			return next(c)
		}
	}
}
