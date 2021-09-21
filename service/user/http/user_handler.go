package http

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	helperModel "git.innovasive.co.th/backend/models"
	"github.com/BlackMocca/go-clean-template/constants"
	"github.com/BlackMocca/go-clean-template/models"
	"github.com/BlackMocca/go-clean-template/service/user"
	"github.com/labstack/echo/v4"
)

type userHandler struct {
	userUs user.UserUsecase
}

func NewUserHandler(e *echo.Echo, us user.UserUsecase) user.UserHandler {
	return &userHandler{
		userUs: us,
	}
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

func (u *userHandler) FetchOneByUserId(c echo.Context) error {
	var userId, err = strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	user, err := u.userUs.FetchOneById(userId)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if user == nil {
		return echo.NewHTTPError(http.StatusNotFound, errors.New("not found").Error())
	}

	responseData := map[string]interface{}{
		"user": user,
	}
	return c.JSON(http.StatusOK, responseData)
}

func (u *userHandler) Create(c echo.Context) error {
	var params = c.Get("params").(map[string]interface{})
	var t = helperModel.NewTimestampFromTime(time.Now())

	user := models.NewUserWithParams(params, nil)
	user.GenUUID()
	user.CreatedAt = &t
	user.UpdatedAt = &t

	if err := u.userUs.Create(user); err != nil {
		if err.Error() == constants.ERROR_DUPLICATE_EMAIL_MESSAGE {
			return echo.NewHTTPError(http.StatusConflict, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	responseData := map[string]interface{}{
		"user": user,
	}
	return c.JSON(http.StatusOK, responseData)
}
