package http

import (
	"net/http"
	"time"

	helperModel "git.innovasive.co.th/backend/models"
	"github.com/BlackMocca/go-clean-template/middleware"
	"github.com/BlackMocca/go-clean-template/models"
	"github.com/BlackMocca/go-clean-template/service/user"
	"github.com/labstack/echo/v4"
)

type userHandler struct {
	userUs user.UserUsecaseInf
}

func NewUserHandler(e *echo.Echo, middL *middleware.GoMiddleware, us user.UserUsecaseInf) {
	handler := &userHandler{
		userUs: us,
	}
	e.GET("/users", handler.FetchAll)
	e.POST("/users", handler.Create)
}

func (u *userHandler) FetchAll(c echo.Context) error {
	users, err := u.userUs.FetchAll()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	responseData := map[string]interface{}{
		"users": users,
	}
	return c.JSON(http.StatusOK, responseData)
}

func (u *userHandler) Create(c echo.Context) error {
	var user = new(models.User)
	if err := c.Bind(user); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	t := helperModel.NewTimestampFromTime(time.Now())

	user.CreatedAt = &t
	user.UpdatedAt = &t

	if err := u.userUs.Create(user); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	responseData := map[string]interface{}{
		"user": user,
	}
	return c.JSON(http.StatusOK, responseData)
}
