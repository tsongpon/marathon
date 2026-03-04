package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/labstack/echo/v5"
)

func APIKeyAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			path := c.Request().URL.Path
			if path == "/ping" || (strings.HasPrefix(path, "/alerts/") && strings.HasSuffix(path, "/ack")) {
				return next(c)
			}
			apiKey := c.QueryParam("api_key")
			expectedKey := os.Getenv("API_KEY")
			if expectedKey == "" {
				return echo.NewHTTPError(http.StatusInternalServerError, "API_KEY not configured")
			}
			if apiKey == "" || apiKey != expectedKey {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid or missing api_key")
			}
			return next(c)
		}
	}
}
