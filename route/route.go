package route

import (
	"github.com/Blackmocca/go-clean-template/middleware"
	"github.com/Blackmocca/go-clean-template/service/v1/organize"
	_organize_validator "github.com/Blackmocca/go-clean-template/service/v1/organize/validator"
	"github.com/labstack/echo/v4"
)

type Route struct {
	e     *echo.Echo
	middl middleware.GoMiddlewareInf
}

func NewRoute(e *echo.Echo, middl middleware.GoMiddlewareInf) *Route {
	return &Route{e: e, middl: middl}
}

func (r Route) RegisterOrganization(handler organize.OrganizeHandler, validation _organize_validator.Validation) {
	r.e.GET("/v1/organizes", handler.FetchAll)
	r.e.GET("/v1/organizes/:org_id", handler.FetchOneById)
	r.e.POST("/v1/organizes", handler.CreateOrg, validation.ValidateCreateOrg)
}
