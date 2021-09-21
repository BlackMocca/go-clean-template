package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	helper "github.com/BlackMocca/go-clean-template/helper/json"
)

// GoMiddleware represent the data-struct for middleware
type GoMiddleware struct {
	// another stuff , may be needed by middleware
}

func (m *GoMiddleware) InputForm(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		var data map[string]interface{}
		reqMethod := c.Request().Method
		Header := c.Request().Header

		if reqMethod == http.MethodPost || reqMethod == http.MethodPut {
			contentType := Header.Get("Content-Type")
			if strings.Contains(contentType, echo.MIMEMultipartForm) {
				form, err := c.MultipartForm()
				if err != nil {
					return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": http.ErrMissingBoundary.Error() + " or has not any parameter"})
				}
				if _, ok := form.Value["data"]; ok {
					jsonData := helper.GetParamsFromJsonData(form.Value)
					if jsonData == nil {
						return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": "Json parsing error"})
					}
					data = jsonData.(map[string]interface{})
				}
				/* รูปสำหรับ ใช้งานทั่วไป */
				if val, ok := form.File["files"]; ok {
					c.Set("files", val)
				}
			} else {
				postForm, err := c.FormParams()
				if err != nil {
					return c.JSON(http.StatusBadRequest, err.Error())
				}
				if len(postForm) > 0 {
					jsonData := helper.GetParamsFromJsonData(postForm)
					if jsonData == nil {
						return c.JSON(http.StatusBadRequest, map[string]interface{}{"message": "Json parsing error"})
					}
					data = jsonData.(map[string]interface{})
				}
			}
		}

		c.Set("params", data)
		return next(c)
	}
}

// InitMiddleware intialize the middleware
func InitMiddleware() *GoMiddleware {
	return &GoMiddleware{}
}
