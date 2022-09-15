package organize

import "github.com/labstack/echo/v4"

type OrganizeHandler interface {
	FetchAll(c echo.Context) error
	FetchOneById(c echo.Context) error
	CreateOrg(c echo.Context) error
}
