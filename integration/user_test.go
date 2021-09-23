// build integration

package integration_test

import (
	"encoding/json"
	"net/http"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

func (e *E2eTestSuite) TestCreateUser_Success() {
	body := `{
		"email": "porn@gmail.com",
		"firstname": "Pornpan",
		"lastname": "Nimnung",
		"age": 25,
		"user_type_id": "ac809526-f9b6-4c4e-a11d-92a3a3f3c117"
	}`

	var bodyM = map[string]interface{}{}
	json.Unmarshal([]byte(body), &bodyM)

	resp, err := resty.New().R().
		SetHeader("Content-Type", "application/json").
		SetBody(bodyM).
		Post("http://127.0.0.1:3000/users")
	if err != nil {
		e.T().Error(err)
	}

	var expectedBody map[string]interface{}
	json.Unmarshal(resp.Body(), &expectedBody)

	assert.NotNil(e.T(), expectedBody["user"])
	var user = expectedBody["user"].(map[string]interface{})

	assert.Equal(e.T(), resp.StatusCode(), http.StatusOK)
	assert.NotNil(e.T(), user["id"])
	assert.Equal(e.T(), user["email"], bodyM["email"])
	assert.Equal(e.T(), user["firstname"], bodyM["firstname"])
	assert.Equal(e.T(), user["lastname"], bodyM["lastname"])
	assert.Equal(e.T(), user["age"], bodyM["age"])
	assert.NotNil(e.T(), user["created_at"])
	assert.NotNil(e.T(), user["updated_at"])
}
