package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"

	"example/constants"

	"github.com/labstack/echo/v4"
)

func (m client) TransformResponse(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if err := next(c); err != nil {
			switch v := err.(type) {
			case Response:
				if v.HttpStatus >= http.StatusInternalServerError {
					ecErr := &echo.HTTPError{
						Code:    v.HttpStatus,
						Message: v,
					}
					c.Error(ecErr)
					return nil
				}
				return c.JSON(v.HttpStatus, v)
			}
			return err
		}
		return nil
	}
}

type Response struct {
	HttpStatus   int                            `json:"-"`
	ResponseCode constants.ApplicationCode      `json:"code"`
	ErrorType    constants.ApplicationErrorType `json:"error_type,omitempty"`
	Message      string                         `json:"message"`
	Data         interface{}                    `json:"data"`
}

func (r Response) Error() string {
	return fmt.Sprintf("code=%d error_type=%s message=%s", r.HttpStatus, r.ErrorType, r.Message)
}

func (r Response) String() string {
	bu, _ := json.Marshal(r)
	return string(bu)
}

func NewResponseJSON(data interface{}) Response {
	return Response{
		HttpStatus:   http.StatusOK,
		ResponseCode: constants.ApplicationCodeSuccess,
		Message:      "successful",
		Data:         data,
	}
}

func NewResponseError(httpStatus int, code constants.ApplicationCode, errorType constants.ApplicationErrorType, message string, data interface{}) Response {
	return Response{
		HttpStatus:   httpStatus,
		ResponseCode: code,
		ErrorType:    errorType,
		Message:      message,
		Data:         data,
	}
}

func NewResponseNoContent() Response {
	return Response{
		HttpStatus: http.StatusNoContent,
	}
}
