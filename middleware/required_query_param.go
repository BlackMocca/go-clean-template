package middleware

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

func (m *GoMiddleware) RequireQueryParam(key string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			param := c.QueryParam(key)
			if param == "" {
				return echo.NewHTTPError(http.StatusBadRequest, errors.New(fmt.Sprintf("%s must be required", key)).Error())
			}
			return next(c)
		}
	}
}
