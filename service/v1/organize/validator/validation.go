package validator

import (
	"fmt"
	"net/http"

	"git.innovasive.co.th/backend/helper"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/labstack/echo/v4"
)

type Validation struct{}

func (v Validation) ValidateCreateOrg(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		var params = c.Get("params").(map[string]interface{})

		/* key params */
		key := "name"
		name, nameOK := params[key]
		if !nameOK {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("%s: was missing on body", key).Error())
		}
		if name != nil && name != "" {
			if err := validation.Validate(name, validation.By(helper.ValidateTypeString)); err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("%s: %s", key, err.Error()).Error())
			}
		}

		key = "alias_name"
		aliasName, aliasNameOK := params[key]
		if !aliasNameOK {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("%s: was missing on body", key).Error())
		}
		if aliasName != nil && aliasName != "" {
			if err := validation.Validate(aliasName, validation.By(helper.ValidateTypeString)); err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("%s: %s", key, err.Error()).Error())
			}
		}

		key = "private_tel_no"
		privateTelNo, privateTelNoOK := params[key]
		if !privateTelNoOK {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("%s: was missing on body", key).Error())
		}
		if privateTelNo != nil && privateTelNo != "" {
			if err := validation.Validate(privateTelNo, validation.By(helper.ValidateTypeString)); err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("%s: %s", key, err.Error()).Error())
			}
		}

		return next(c)
	}
}
