package auth

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

// revokeTokenRequest is the request body for POST /auth/revoke.
type revokeTokenRequest struct {
	JTI       string    `json:"jti"`
	ExpiresAt time.Time `json:"expires_at"`
	UserID    string    `json:"user_id,omitempty"`
}

// revokeUserRequest is the request body for POST /auth/revoke-user.
type revokeUserRequest struct {
	UserID string `json:"user_id"`
}

// revokeUserResponse is the response for POST /auth/revoke-user.
type revokeUserResponse struct {
	RevokedCount int `json:"revoked_count"`
}

// revocationListResponse is the response for GET /auth/revocations.
type revocationListResponse struct {
	Count   int              `json:"count"`
	Entries []RevocationInfo `json:"entries"`
}

// RegisterRevocationRoutes registers token revocation management endpoints.
// All endpoints require the "admin" role.
func RegisterRevocationRoutes(g *echo.Group, store *TokenRevocationStore) {
	authGroup := g.Group("/auth", RequireRole("admin"))

	authGroup.POST("/revoke", handleRevokeToken(store))
	authGroup.POST("/revoke-user", handleRevokeUser(store))
	authGroup.GET("/revocations", handleListRevocations(store))
}

// handleRevokeToken revokes a specific token by JTI.
func handleRevokeToken(store *TokenRevocationStore) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req revokeTokenRequest
		if err := c.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
		}

		if req.JTI == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "jti is required")
		}

		if req.ExpiresAt.IsZero() {
			// Default to 1 hour from now if no expiry provided
			req.ExpiresAt = time.Now().Add(1 * time.Hour)
		}

		if req.UserID != "" {
			store.RevokeForUser(req.JTI, req.UserID, req.ExpiresAt)
		} else {
			store.Revoke(req.JTI, req.ExpiresAt)
		}

		return c.NoContent(http.StatusNoContent)
	}
}

// handleRevokeUser revokes all known tokens for a user.
func handleRevokeUser(store *TokenRevocationStore) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req revokeUserRequest
		if err := c.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
		}

		if req.UserID == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "user_id is required")
		}

		count := store.RevokeAllForUser(req.UserID)
		return c.JSON(http.StatusOK, revokeUserResponse{RevokedCount: count})
	}
}

// handleListRevocations returns all currently active revocation entries.
func handleListRevocations(store *TokenRevocationStore) echo.HandlerFunc {
	return func(c echo.Context) error {
		entries := store.Entries()
		return c.JSON(http.StatusOK, revocationListResponse{
			Count:   len(entries),
			Entries: entries,
		})
	}
}
