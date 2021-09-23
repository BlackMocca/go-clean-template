package user

import "github.com/labstack/echo/v4"

type UserHandler interface {
	FetchAll(c echo.Context) error
	FetchOneByUserId(c echo.Context) error
	Create(c echo.Context) error
}
