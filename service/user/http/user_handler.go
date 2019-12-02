package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"gitlab.com/km/go-kafka-playground/middleware"
	"gitlab.com/km/go-kafka-playground/service/user"
)

type userHandler struct {
	userUs user.UserUsecaseInf
}

func NewUserHandler(e *echo.Echo, middL *middleware.GoMiddleware, us user.UserUsecaseInf) {
	handler := &userHandler{
		userUs: us,
	}
	e.GET("/users", handler.Create)
}

func (u *userHandler) Create(c echo.Context) error {
	responseData := map[string]interface{}{
		"test": "test",
	}
	return c.JSON(http.StatusOK, responseData)
}
