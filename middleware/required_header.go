package middleware

import (
	"fmt"
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/labstack/echo/v4"
)

func (m *GoMiddleware) ValidateRequiredHeader(key string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			headerVal := c.Request().Header.Get(key)
			if err := validation.Validate(headerVal, validation.Required); err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, fmt.Errorf("header %s: %s", key, err.Error()).Error())
			}

			return next(c)
		}
	}
}
