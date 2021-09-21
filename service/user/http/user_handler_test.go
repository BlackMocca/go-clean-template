package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	helperModel "git.innovasive.co.th/backend/models"
	"github.com/BlackMocca/go-clean-template/models"
	"github.com/BlackMocca/go-clean-template/service/user/mocks"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetUserById(t *testing.T) {
	createdAt := helperModel.NewTimestampFromString("2020-09-22 03:43:01")
	updatedAt := helperModel.NewTimestampFromString("2020-09-22 03:43:01")
	mockUser := models.User{
		ID:        1,
		Email:     "teeradet.huag@gmail.com",
		Firstname: "ธีรเดช",
		Lastname:  "พลเดชปริญญา",
		Age:       50,
		CreatedAt: &createdAt,
		UpdatedAt: &updatedAt,
	}

	mockUs := new(mocks.UserUsecaseInf)
	// Setup
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/users/:id")
	c.SetParamNames("id")
	c.SetParamValues("1")
	// Assertions
	t.Run("success=200", func(t *testing.T) {
		handler := userHandler{userUs: mockUs}
		mockUs.On("FetchOneById", mock.AnythingOfType("int64")).Return(&mockUser, nil)
		if assert.NoError(t, handler.FetchOneByUserId(c)) {
			assertString := `{"user":{"id":1,"email":"teeradet.huag@gmail.com","firstname":"ธีรเดช","lastname":"พลเดชปริญญา","age":50,"created_at":"2020-09-22 03:43:01","updated_at":"2020-09-22 03:43:01","deleted_at":null}}`

			assertData := make(map[string]interface{})
			respData := make(map[string]interface{})
			err := json.Unmarshal([]byte(assertString), &assertData)
			err = json.Unmarshal([]byte(rec.Body.String()), &respData)
			assert.NoError(t, err)

			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, assertData, respData)
		}
	})
	/* case error 500 */
	// t.Run("error=500", func(t *testing.T) {
	// 	handler := userHandler{userUs: mockUs}
	// 	mockUs.On("FetchOneById", mock.AnythingOfType("int64")).Return(nil, errors.New("Unexpected"))
	// 	err := handler.FetchOneByUserId(c)
	// 	echoErr := err.(*echo.HTTPError)
	// 	assert.Equal(t, http.StatusInternalServerError, echoErr.Code)

	// 	/* assert message */
	// 	respJSON, _ := json.Marshal(echoErr)
	// 	assertString := `{"message":"Unexpected"}`
	// 	assertData := make(map[string]interface{})
	// 	respData := make(map[string]interface{})
	// 	err = json.Unmarshal([]byte(assertString), &assertData)
	// 	err = json.Unmarshal([]byte(respJSON), &respData)

	// 	assert.NoError(t, err)
	// 	assert.Equal(t, assertData, respData)
	// })
}
