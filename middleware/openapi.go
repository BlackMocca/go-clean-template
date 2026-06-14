package middleware

import (
	"context"
	"net/http"
	"strings"

	"example/constants"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/labstack/echo/v4"
	middleware "github.com/oapi-codegen/echo-middleware"
	"github.com/spf13/cast"
)

func (m client) CreateOpenapiMiddleware(getSwaggers ...func() (*openapi3.T, error)) echo.MiddlewareFunc {
	validators := make([]echo.MiddlewareFunc, 0, len(getSwaggers))

	for _, getSwagger := range getSwaggers {
		spec, err := getSwagger()
		if err != nil {
			continue
		}
		// skip auth in apispec using custom in route instead
		validator := middleware.OapiRequestValidatorWithOptions(spec, &middleware.Options{
			Options: openapi3filter.Options{
				AuthenticationFunc: func(ctx context.Context, ai *openapi3filter.AuthenticationInput) error { return nil },
			},
			ErrorHandler: func(c echo.Context, err *echo.HTTPError) error {
				if err != nil && err.Code == http.StatusNotFound {
					return NewResponseError(http.StatusNotFound, constants.ApplicationCodeNotFound, constants.ApplicationErrorNotFound, cast.ToString(err.Message), nil)
				}

				return NewResponseError(http.StatusBadRequest, constants.ApplicationCodeBadRequestInvalidArgument, constants.ApplicationErrorInvalidArgument, cast.ToString(err.Message), nil)
			},
			SilenceServersWarning: true,
		})
		validators = append(validators, validator)
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			var matched bool
			var epError error

			for _, v := range validators {
				// ใช้ middleware ตรวจเฉย ๆ ด้วย dummy handler
				err := v(func(c echo.Context) error {
					matched = true
					return nil
				})(c)

				epError = err

				if err == nil && matched {
					break
				}

				if !strings.Contains(epError.Error(), "no matching operation was found") {
					return NewResponseError(http.StatusBadRequest, constants.ApplicationCodeBadRequestInvalidArgument, constants.ApplicationErrorNotFound, epError.Error(), nil)
				}

			}

			if !matched && epError != nil {
				if !strings.Contains(epError.Error(), "no matching operation was found") {
					return NewResponseError(http.StatusBadRequest, constants.ApplicationCodeBadRequestInvalidArgument, constants.ApplicationErrorNotFound, epError.Error(), nil)
				}
			}

			return next(c)
		}
	}
}
