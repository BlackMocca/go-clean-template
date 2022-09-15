package middleware

import (
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
)

func (m *GoMiddleware) ValidateParamId(key string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			param := c.Param(key)
			_, err := uuid.FromString(param)
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, err.Error())
			}
			return next(c)
		}
	}
}
