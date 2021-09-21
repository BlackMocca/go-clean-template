package middleware

import (
	"github.com/labstack/echo/v4"
)

func (m *GoMiddleware) IsAuthorization(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {

		if err := IsAuthorization(c); err != nil {
			var code int
			var message interface{}
			if he, ok := err.(*echo.HTTPError); ok {
				code = he.Code
				message = he.Message
			}
			return echo.NewHTTPError(code, message)
		}
		return next(c)
	}
}

func IsAuthorization(c echo.Context) error {
	var header = c.Request().Header
	if len(header) == 0 {
		return echo.NewHTTPError(401, map[string]interface{}{"error": "Unauthorize"})
	}

	authorization := header.Get("Authorization")
	if authorization == "" {
		return echo.NewHTTPError(401, map[string]interface{}{"error": "Unauthorize"})
	}

	return nil
}
