package oauthsession

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

var (
	// ErrSessionNotFound means there is no row or no stored Discord access token for the user.
	ErrSessionNotFound = echo.NewHTTPError(http.StatusUnauthorized, "Session not found; please re-authenticate.")
	// ErrDiscordTokenInvalid means the stored Discord token could not be refreshed (e.g. revoked).
	ErrDiscordTokenInvalid = echo.NewHTTPError(http.StatusUnauthorized, "Discord session expired; please re-authenticate.")
)
