package route

import (
	"github.com/BlackMocca/go-clean-template/middleware"
	"github.com/BlackMocca/go-clean-template/service/user"
	_user_validator "github.com/BlackMocca/go-clean-template/service/user/validator"
	"github.com/labstack/echo/v4"
)

type Route struct {
	e     *echo.Echo
	middl middleware.GoMiddlewareInf
}

func NewRoute(e *echo.Echo, middl middleware.GoMiddlewareInf) *Route {
	return &Route{e: e, middl: middl}
}

func (r Route) RegisterRouteUser(handler user.UserHandler, validation _user_validator.Validation) {
	r.e.GET("/users", handler.FetchAll)
	r.e.GET("/users/:id", handler.FetchOneByUserId)
	r.e.POST("/users", handler.Create)
}
