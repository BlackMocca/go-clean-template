package middleware

import (
	"context"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"
)

// Middleware defines the interface for authentication middleware that Echo routes can use.
// Implementations validate requests and reject unauthorized access.
type Middleware interface {
	// APIKeyAuth wraps an echo handler with API key validation logic.
	// Return this wrapped handler when registering protected routes.
	APIKeyAuth(next echo.HandlerFunc) echo.HandlerFunc
	// response format following openapispec
	TransformResponse(next echo.HandlerFunc) echo.HandlerFunc
	InitContextIfNotExists(next echo.HandlerFunc) echo.HandlerFunc
	CreateOpenapiMiddleware(getSwaggers ...func() (*openapi3.T, error)) echo.MiddlewareFunc
}

// client is the internal implementation of the Middleware interface.
// It holds the APIKeyConfig that determines validation behavior.
type client struct {
	apikeyConfig apiKeyConfig
}

// NewClient creates a Middleware that uses the provided APIKeyConfig for authentication.
func NewClient(apikeyConfig apiKeyConfig) Middleware {
	return &client{apikeyConfig: apikeyConfig}
}

func (c client) InitContextIfNotExists(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		if ctx == nil {
			bgCtx := context.Background()
			newReq := c.Request().WithContext(bgCtx)

			c.SetRequest(newReq)
		}
		return next(c)
	}
}
