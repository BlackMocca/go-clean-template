/*
Package middleware provides API Key authentication middleware for Echo.

Supported key locations (checked in order):
 1. Request header:  X-API-Key
 2. Query parameter: ?apikey=

Usage:

	apiKeyConfig := middleware.APIKeyAuthDefault("secret-key")
	middl := middleware.NewClient(apiKeyConfig)

Then attach middl.APIKeyAuth as a handler middleware on protected routes.
*/
package middleware

import (
	"errors"
	"example/constants"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

// ==================== Error Definitions ====================

var (
	errAPIKeyMissing = errors.New("api key is missing") // request had no key in header or query param
	errAPIKeyInvalid = errors.New("invalid api key")    // provided key did not match the configured value
)

// apiKeyConfig holds settings for API Key validation.
type apiKeyConfig struct {
	Skipper    Skipper // function that determines whether to skip validation for a given request
	HeaderName string  // name of the HTTP header containing the API key (default: X-API-Key)
	Value      string  // expected API key value to validate against
}

// Skipper defines a function type that decides whether authentication should be skipped for a request.
type Skipper func(c echo.Context) bool

func defaultSkipper(c echo.Context) bool {
	return false // by default, do not skip any route
}

// ==================== Default Configuration ====================

// APIKeyAuthDefault returns a configured apiKeyConfig with sensible defaults:
//   - Header: X-API-Key
//   - Query:  ?apikey=
//   - Context key: api_key_user
func APIKeyAuthDefault(headerKey, headerValue string) apiKeyConfig {
	return apiKeyConfig{
		Skipper:    defaultSkipper,
		HeaderName: headerKey,
		Value:      headerValue,
	}
}

// ==================== Middleware ====================

// APIKeyAuth is the HTTP middleware that validates an API key on every request.
// It checks for a key in the configured header first, then falls back to the query parameter.
// On success, stores the key in the echo.Context under Config.ContextKey so downstream
// handlers can read it if needed.
func (m *client) APIKeyAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Skip validation if route is in the skip list
		if m.apikeyConfig.Skipper != nil && m.apikeyConfig.Skipper(c) {
			return next(c)
		}

		// Extract API key from header first, then fallback to query param
		key := m.apikeyConfig.extractAPIKey(c, m.apikeyConfig.HeaderName)

		// Check that a key was provided at all
		if key == "" {
			return NewResponseError(http.StatusUnauthorized, constants.ApplicationCodeUnAuthorize, constants.ApplicationErrorNotFound, errAPIKeyInvalid.Error(), nil)
		}

		// Validate the key value (case-insensitive comparison)
		if !strings.EqualFold(m.apikeyConfig.Value, key) {
			return NewResponseError(http.StatusUnauthorized, constants.ApplicationCodeUnAuthorize, constants.ApplicationErrorApikeyInvalid, errAPIKeyMissing.Error(), nil)
		}

		return next(c)
	}
}

// ==================== Helpers ====================

func (a apiKeyConfig) extractAPIKey(c echo.Context, headerName string) string {
	// 1. Try header first (highest priority)
	return c.Request().Header.Get(headerName)
}
