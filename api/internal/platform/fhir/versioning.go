package fhir

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

// VersionedResource is implemented by domain models that support versioning.
type VersionedResource interface {
	GetVersionID() int
	SetVersionID(v int)
}

// SetVersionHeaders sets ETag and Last-Modified headers on the response.
func SetVersionHeaders(c echo.Context, versionID int, lastModified string) {
	c.Response().Header().Set("ETag", fmt.Sprintf(`W/"%d"`, versionID))
	if lastModified != "" {
		c.Response().Header().Set("Last-Modified", lastModified)
	}
}

// CheckIfMatch validates the If-Match header against the current version.
// Returns 0, nil if no If-Match header is present (unconditional update).
// Returns the expected version if header is present and valid.
// Returns an error response if versions don't match (409 Conflict).
func CheckIfMatch(c echo.Context, currentVersion int) (int, error) {
	ifMatch := c.Request().Header.Get("If-Match")
	if ifMatch == "" {
		return 0, nil
	}

	expectedVersion, err := ParseETag(ifMatch)
	if err != nil {
		return 0, echo.NewHTTPError(http.StatusBadRequest, "invalid If-Match header: "+err.Error())
	}

	if expectedVersion != currentVersion {
		return 0, echo.NewHTTPError(http.StatusConflict,
			fmt.Sprintf("version conflict: expected version %d but resource is at version %d", expectedVersion, currentVersion))
	}

	return expectedVersion, nil
}

// ParseETag extracts the version number from an ETag value like W/"3" or "3".
func ParseETag(etag string) (int, error) {
	etag = strings.TrimSpace(etag)
	// Remove weak indicator
	etag = strings.TrimPrefix(etag, "W/")
	// Remove quotes
	etag = strings.Trim(etag, `"`)

	v, err := strconv.Atoi(etag)
	if err != nil {
		return 0, fmt.Errorf("ETag must contain a numeric version: %s", etag)
	}
	return v, nil
}

// FormatETag creates a weak ETag from a version ID.
func FormatETag(versionID int) string {
	return fmt.Sprintf(`W/"%d"`, versionID)
}

// CheckIfNoneMatch checks If-None-Match for conditional reads.
// Returns true if the client's version matches (304 Not Modified should be returned).
func CheckIfNoneMatch(c echo.Context, currentVersion int) bool {
	ifNoneMatch := c.Request().Header.Get("If-None-Match")
	if ifNoneMatch == "" {
		return false
	}

	clientVersion, err := ParseETag(ifNoneMatch)
	if err != nil {
		return false
	}

	return clientVersion == currentVersion
}
