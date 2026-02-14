package db

import (
	"context"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

// PoolStats represents database connection pool statistics.
type PoolStats struct {
	TotalConns      int32  `json:"total_conns"`
	IdleConns       int32  `json:"idle_conns"`
	AcquiredConns   int32  `json:"acquired_conns"`
	MaxConns        int32  `json:"max_conns"`
	AcquireCount    int64  `json:"acquire_count"`
	AcquireDuration string `json:"acquire_duration"`
	Healthy         bool   `json:"healthy"`
}

// GetPoolStats returns connection pool statistics.
func GetPoolStats(pool *pgxpool.Pool) *PoolStats {
	stat := pool.Stat()
	return &PoolStats{
		TotalConns:      stat.TotalConns(),
		IdleConns:       stat.IdleConns(),
		AcquiredConns:   stat.AcquiredConns(),
		MaxConns:        stat.MaxConns(),
		AcquireCount:    stat.AcquireCount(),
		AcquireDuration: stat.AcquireDuration().String(),
		Healthy:         stat.TotalConns() > 0,
	}
}

// HealthHandler returns a handler for the database health check endpoint.
func HealthHandler(pool *pgxpool.Pool) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
		defer cancel()

		// Ping the database
		err := pool.Ping(ctx)
		stats := GetPoolStats(pool)

		if err != nil {
			stats.Healthy = false
			return c.JSON(http.StatusServiceUnavailable, map[string]interface{}{
				"status": "unhealthy",
				"error":  err.Error(),
				"pool":   stats,
			})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"status": "healthy",
			"pool":   stats,
		})
	}
}
