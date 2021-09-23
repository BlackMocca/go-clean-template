package validator

import (
	"fmt"
	"net/http"

	"git.innovasive.co.th/backend/helper"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	rule "github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"
)

type Validation struct {
}

func (v Validation) ValidateCreateUser(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		var params = c.Get("params").(map[string]interface{})
		var key string

		/* email */
		key = "email"
		v, ok := params[key]
		if !ok {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("%s: %s", key, "not found on body"))
		}
		if err := validation.Validate(v, validation.By(helper.ValidateTypeString), rule.EmailFormat); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("%s: %s", key, err.Error()))
		}

		/* firstname */
		key = "firstname"
		firstname, ok := params[key]
		if !ok {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("%s: %s", key, "not found on body"))
		}
		if err := validation.Validate(firstname, validation.By(helper.ValidateTypeString)); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("%s: %s", key, err.Error()))
		}

		/* lastname */
		key = "lastname"
		lastname, ok := params[key]
		if !ok {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("%s: %s", key, "not found on body"))
		}
		if err := validation.Validate(lastname, validation.By(helper.ValidateTypeString)); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("%s: %s", key, err.Error()))
		}

		/* age */
		key = "age"
		age, ok := params[key]
		if !ok {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("%s: %s", key, "not found on body"))
		}
		if err := validation.Validate(age, validation.By(helper.ValidateTypeInt)); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("%s: %s", key, err.Error()))
		}

		return next(c)
	}
}
