package oauthsession

import (
	"github.com/labstack/echo/v4"
)

var (
	// ErrSessionNotFound means there is no row or no stored Discord access token for the user.
	ErrSessionNotFound = echo.NewHTTPError(401, "Session not found; please re-authenticate.")
	// ErrDiscordTokenInvalid means the stored Discord token could not be refreshed (e.g. revoked).
	ErrDiscordTokenInvalid = echo.NewHTTPError(401, "Discord session expired; please re-authenticate.")
)
