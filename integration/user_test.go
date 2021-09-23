// build integration

package integration_test

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/go-resty/resty/v2"
	"github.com/gofrs/uuid"
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
	assert.Equal(e.T(), user["user_type_id"], bodyM["user_type_id"])
	assert.NotNil(e.T(), user["created_at"])
	assert.NotNil(e.T(), user["updated_at"])
}

func (e *E2eTestSuite) TestFetchAll_WITH_NO_ARGS_Success_MIGRATE_STORY_001() {
	resp, err := resty.New().R().
		Get("http://127.0.0.1:3000/users")
	if err != nil {
		e.T().Error(err)
	}

	var expectedBody map[string]interface{}
	json.Unmarshal(resp.Body(), &expectedBody)

	assert.NotNil(e.T(), expectedBody["users"])
	var users = expectedBody["users"].([]interface{})

	assert.Equal(e.T(), resp.StatusCode(), http.StatusOK)
	assert.NotEmpty(e.T(), users)

	for _, userM := range users {
		user := userM.(map[string]interface{})
		assert.NotNil(e.T(), user["id"])
		assert.NotZero(e.T(), user["email"])
		assert.NotZero(e.T(), user["firstname"])
		assert.NotZero(e.T(), user["lastname"])
		assert.NotZero(e.T(), user["age"])
		assert.NotZero(e.T(), user["user_type_id"])
		assert.NotNil(e.T(), user["user_type"])
		assert.NotNil(e.T(), user["created_at"])
		assert.NotNil(e.T(), user["updated_at"])
	}
}

func (e *E2eTestSuite) TestFetchAll_WITH_USER_TYPE_Success_MIGRATE_STORY_001() {
	userTypeId := "8c22b186-ae36-481a-b06d-6fe7a3a48a77"
	queryParams := url.Values{}
	queryParams.Add("user_type_id", userTypeId)

	resp, err := resty.New().R().
		SetQueryString(queryParams.Encode()).
		Get("http://127.0.0.1:3000/users")
	if err != nil {
		e.T().Error(err)
	}

	var expectedBody map[string]interface{}
	json.Unmarshal(resp.Body(), &expectedBody)

	assert.NotNil(e.T(), expectedBody["users"])
	var users = expectedBody["users"].([]interface{})

	assert.Equal(e.T(), resp.StatusCode(), http.StatusOK)
	assert.NotEmpty(e.T(), users)

	for _, userM := range users {
		user := userM.(map[string]interface{})
		assert.NotNil(e.T(), user["id"])
		assert.NotZero(e.T(), user["email"])
		assert.NotZero(e.T(), user["firstname"])
		assert.NotZero(e.T(), user["lastname"])
		assert.NotZero(e.T(), user["age"])
		assert.NotZero(e.T(), user["user_type_id"])
		assert.Equal(e.T(), user["user_type_id"], userTypeId)
		assert.NotNil(e.T(), user["user_type"])
		assert.NotNil(e.T(), user["created_at"])
		assert.NotNil(e.T(), user["updated_at"])
	}
}

func (e *E2eTestSuite) TestFetchAll_SERCH_NOT_FOUND_Success_MIGRATE_STORY_001() {
	userTypeId := "9258f398-53da-4d37-ab97-e76df151a150"
	queryParams := url.Values{}
	queryParams.Add("user_type_id", userTypeId)

	resp, err := resty.New().R().
		SetQueryString(queryParams.Encode()).
		Get("http://127.0.0.1:3000/users")
	if err != nil {
		e.T().Error(err)
	}

	assert.Equal(e.T(), resp.StatusCode(), http.StatusNoContent)
}

func (e *E2eTestSuite) TestFetchOneById_Success_MIGRATE_STORY_001() {
	id := uuid.FromStringOrNil("93a42c6c-d52b-42a6-a33e-31ab74b7c567")
	resp, err := resty.New().R().
		SetPathParams(map[string]string{
			"user_id": id.String(),
		}).
		Get("http://127.0.0.1:3000/users/{user_id}")
	if err != nil {
		e.T().Error(err)
	}

	var expectedBody map[string]interface{}
	json.Unmarshal(resp.Body(), &expectedBody)

	assert.NotNil(e.T(), expectedBody["user"])
	var user = expectedBody["user"].(map[string]interface{})

	assert.Equal(e.T(), resp.StatusCode(), http.StatusOK)
	assert.NotNil(e.T(), user["id"])
	assert.NotZero(e.T(), user["email"])
	assert.NotZero(e.T(), user["firstname"])
	assert.NotZero(e.T(), user["lastname"])
	assert.NotZero(e.T(), user["age"])
	assert.NotZero(e.T(), user["user_type_id"])
	assert.NotNil(e.T(), user["user_type"])
	assert.NotNil(e.T(), user["created_at"])
	assert.NotNil(e.T(), user["updated_at"])
}

func (e *E2eTestSuite) TestFetchOneById_NO_CONTENT() {
	id := uuid.FromStringOrNil("93a42c6c-d52b-42a6-a33e-31ab74b7c567")
	resp, err := resty.New().R().
		SetPathParams(map[string]string{
			"user_id": id.String(),
		}).
		Get("http://127.0.0.1:3000/users/{user_id}")
	if err != nil {
		e.T().Error(err)
	}

	assert.Equal(e.T(), resp.StatusCode(), http.StatusNoContent)
}
