// Package route defines HTTP route registration for the example API server.
// It separates public endpoints (no auth required) from protected ones (API key required).
package route

import (
	"example/middleware"
	"example/service/web"

	"github.com/labstack/echo/v4"
)

// Route encapsulates an Echo router and authentication middleware for unified registration.
type Route struct {
	e     *echo.Echo            // the Echo HTTP server instance
	middl middleware.Middleware // used to apply API key auth to protected routes
}

// NewRoute creates a new Route bound to the given Echo instance and middleware.
func NewRoute(e *echo.Echo, middl middleware.Middleware) *Route {
	return &Route{
		e:     e,
		middl: middl,
	}
}

// RegisterPublicRoutes attaches endpoints that do not require authentication.
// These routes are accessible without an API key.
func (r Route) RegisterPublicRoutes(w web.ServerInterfaceWrapper) {
	// GET /api/health — liveness probe for load balancers and orchestrators
	r.e.GET("/api/health", w.HealthCheck)
}

// RegisterProtectedRoutes attaches endpoints that require a valid API key.
// The middl.APIKeyAuth middleware validates the key on every request.
func (r Route) RegisterProtectedRoutes(w web.ServerInterfaceWrapper) {
	// ===== Users endpoints =====

	// GET /api/v1/users — list all users with optional pagination (page, page_size params)
	r.e.GET("/api/v1/users", w.ListUsers, r.middl.APIKeyAuth)

	// POST /api/v1/users — create a new user; body: { email, name, role }
	r.e.POST("/api/v1/users", w.CreateUser, r.middl.APIKeyAuth)

	// GET /api/v1/users/:id — retrieve a single user by ID
	r.e.GET("/api/v1/users/:id", w.GetUserByID, r.middl.APIKeyAuth)

	// PATCH /api/v1/users/:id — partially update an existing user
	r.e.PUT("/api/v1/users/:id", w.UpdateUserByID, r.middl.APIKeyAuth)

	// DELETE /api/v1/users/:id — permanently remove a user
	r.e.DELETE("/api/v1/users/:id", w.DeleteUserByID, r.middl.APIKeyAuth)

}
