package web_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"example/constants"
	"example/middleware"
	"example/models"
	"example/service/user"
	"example/service/web"
	mocks "example/service/user/mocks"

	utilModels "github.com/BlackMocca/utils/models"
	"github.com/google/uuid"
	"github.com/guregu/null/v6/zero"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// ===========================================================================
// Helpers
// ===========================================================================
func zeroString(s string) zero.String { return zero.StringFrom(s) }

func buildUser(id uuid.UUID, email string, role constants.Role) *models.User {
	now := utilModels.NewTimestampFromNow()
	return &models.User{
		ID:        id,
		Name:      zeroString("Test User"),
		Email:     zeroString(email),
		Role:      utilModels.NewEnumScan[constants.Role](string(role)),
		CreatedAt: now.ToPointer(),
		UpdatedAt: now.ToPointer(),
	}
}

// newRequest creates an http.Request with optional JSON body.
func newRequest(method, url string, bodyJSON ...[]byte) *http.Request {
	if len(bodyJSON) > 0 && bodyJSON[0] != nil {
		r := httptest.NewRequest(method, url, bytes.NewReader(bodyJSON[0]))
		r.Header.Set("Content-Type", "application/json")
		return r
	}
	return httptest.NewRequest(method, url, nil)
}

// extractResponse asserts that the returned error is a middleware.Response (not a real error),
// then returns the Response for further assertions.
func extractResponse(t *testing.T, handlerErr error) middleware.Response {
	t.Helper()
	resp, ok := handlerErr.(middleware.Response)
	require.True(t, ok, "expected middleware.Response but got %T: %v", handlerErr, handlerErr)
	return resp
}

// assertFullResponse checks HTTP status and every field of middleware.Response.
func assertFullResponse(t *testing.T, resp middleware.Response, httpStatus int, code constants.ApplicationCode, errType constants.ApplicationErrorType, message string, data interface{}) {
	assert.Equal(t, httpStatus, resp.HttpStatus)
	assert.Equal(t, code, resp.ResponseCode)
	if errType != "" {
		assert.Equal(t, errType, resp.ErrorType)
	} else {
		assert.Empty(t, string(resp.ErrorType))
	}
	assert.Equal(t, message, resp.Message)
	assert.Equal(t, data, resp.Data)
}

// assertSuccess checks the four core fields of a successful Response (no Data).
func assertSuccessResponse(t *testing.T, resp middleware.Response, httpStatus int, code constants.ApplicationCode, msg string) {
	assertFullResponse(t, resp, httpStatus, code, "", msg, nil)
}

// ===========================================================================
// HealthCheck
// ===========================================================================
func TestHandler_HealthCheck(t *testing.T) {
	mockUc := mocks.NewMockUsecase(t)
	e := echo.New()
	handlerInst := web.NewHandler(mockUc)

	req := newRequest(http.MethodGet, "/api/health")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handlerErr := handlerInst.HealthCheck(c)

	resp := extractResponse(t, handlerErr)

	assertSuccessResponse(t, resp, http.StatusOK, constants.ApplicationCodeSuccess, "successful")
}

// ===========================================================================
// ListUsers — GET /api/v1/users?page=1&page_size=10
// ===========================================================================
func TestHandler_ListUsers(t *testing.T) {
	t.Run("success: returns paginated users with all response fields", func(t *testing.T) {
		mockUc := mocks.NewMockUsecase(t)
		e := echo.New()
		handlerInst := web.NewHandler(mockUc)

		testUsers := []*models.User{
			buildUser(uuid.MustParse("aaaa1111-2222-3333-4444-555555555555"), "alice@example.com", constants.RoleAdmin),
			buildUser(uuid.MustParse("bbbb2222-3333-4444-5555-666666666666"), "bob@example.com", constants.RoleUser),
		}

		paginator := models.Paginator[[]*models.User]{
			Page:          1,
			PageSize:      10,
			TotalDocument: 2,
			TotalPage:     1,
			Documents:     testUsers,
		}

		mockUc.EXPECT().
			ListUsers(mock.Anything, mock.Anything).
			Run(func(ctx context.Context, params user.ListParams) {
				assert.NotNil(t, ctx)
				assert.Equal(t, int64(1), params.Page)
				assert.Equal(t, int64(10), params.PageSize)
				assert.Empty(t, params.Searchword)
				var emptyRole constants.Role
				assert.Equal(t, emptyRole, params.Role)
			}).
			Return(paginator, nil).Times(1)

		req := newRequest(http.MethodGet, "/api/v1/users?page=1&page_size=10")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handlerErr := handlerInst.ListUsers(c, web.ListUsersParams{Page: 1, PageSize: 10})

		resp := extractResponse(t, handlerErr)

		// Success response — core fields (Data is map[string]interface{} from fn.CopyJSON)
		assert.Equal(t, http.StatusOK, resp.HttpStatus)
		assert.Equal(t, constants.ApplicationCodeSuccess, resp.ResponseCode)
		assert.Empty(t, string(resp.ErrorType))
		assert.Equal(t, "successful", resp.Message)
	})

	t.Run("success: returns no-content when documents are empty", func(t *testing.T) {
		mockUc := mocks.NewMockUsecase(t)
		e := echo.New()
		handlerInst := web.NewHandler(mockUc)

		mockUc.EXPECT().
			ListUsers(mock.Anything, mock.Anything).
			Return(models.Paginator[[]*models.User]{
				Page:          1,
				PageSize:      10,
				TotalDocument: 0,
				TotalPage:     0,
				Documents:     []*models.User{},
			}, nil).Times(1)

		req := newRequest(http.MethodGet, "/api/v1/users?page=1&page_size=10")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handlerErr := handlerInst.ListUsers(c, web.ListUsersParams{Page: 1, PageSize: 10})

		resp := extractResponse(t, handlerErr)

		// NoContent response — ResponseCode is empty (not C20000)
		assert.Equal(t, http.StatusNoContent, resp.HttpStatus)
		assert.Empty(t, resp.ResponseCode)
		mockUc.AssertNumberOfCalls(t, "ListUsers", 1)
	})

	t.Run("failure: returns 500 when usecase returns error", func(t *testing.T) {
		mockUc := mocks.NewMockUsecase(t)
		e := echo.New()
		handlerInst := web.NewHandler(mockUc)

		expectedErr := errors.New("database connection refused")

		mockUc.EXPECT().
			ListUsers(mock.Anything, mock.Anything).
			Return(models.Paginator[[]*models.User]{}, expectedErr).Times(1)

		req := newRequest(http.MethodGet, "/api/v1/users")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handlerErr := handlerInst.ListUsers(c, web.ListUsersParams{})

		resp := extractResponse(t, handlerErr)

		assertFullResponse(t, resp, http.StatusInternalServerError, constants.ApplicationInternalServerError, constants.ApplicationErrorInternalServerError, "database connection refused", nil)
		mockUc.AssertNumberOfCalls(t, "ListUsers", 1)
	})
}

// ===========================================================================
// CreateUser — POST /api/v1/users
// ===========================================================================
func TestHandler_CreateUser(t *testing.T) {
	t.Run("success: creates user and returns all response fields", func(t *testing.T) {
		mockUc := mocks.NewMockUsecase(t)
		e := echo.New()
		handlerInst := web.NewHandler(mockUc)

		bodyBytes := []byte(`{"name":"New User","email":"newuser@example.com","role":"user","permission_ids":[]}`)
		mockUc.EXPECT().
			Save(mock.Anything, mock.Anything).
			Run(func(ctx context.Context, u *models.User) {
				assert.NotNil(t, ctx)
				assert.NotEmpty(t, u.ID)
				assert.Equal(t, "New User", u.Name.ValueOrZero())
				assert.Equal(t, "newuser@example.com", u.Email.ValueOrZero())
			}).
			Return(nil).Times(1)

		req := newRequest(http.MethodPost, "/api/v1/users", bodyBytes)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handlerErr := handlerInst.CreateUser(c)

		resp := extractResponse(t, handlerErr)

		// Success response — core fields only (Data is web.User from fn.CopyJSON with typed Data field)
		assert.Equal(t, http.StatusOK, resp.HttpStatus)
		assert.Equal(t, constants.ApplicationCodeSuccess, resp.ResponseCode)
		assert.Empty(t, string(resp.ErrorType))
		assert.Equal(t, "successful", resp.Message)

		mockUc.AssertNumberOfCalls(t, "Save", 1)
	})

	t.Run("failure: returns 409 when email is duplicated", func(t *testing.T) {
		mockUc := mocks.NewMockUsecase(t)
		e := echo.New()
		handlerInst := web.NewHandler(mockUc)

		conflictErr := errors.New(string(constants.ApplicationCodeConflictUserEmailWasDuplicate))

		mockUc.EXPECT().
			Save(mock.Anything, mock.Anything).
			Return(conflictErr).Times(1)

		bodyBytes := []byte(`{"name":"Taken User","email":"taken@example.com","role":"user","permission_ids":[]}`)
		req := newRequest(http.MethodPost, "/api/v1/users", bodyBytes)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handlerErr := handlerInst.CreateUser(c)

		resp := extractResponse(t, handlerErr)

		assertFullResponse(t, resp, http.StatusConflict, constants.ApplicationCodeConflictUserEmailWasDuplicate, constants.ApplicationErrorConflict, "email taken@example.com was duplicated", nil)
		mockUc.AssertNumberOfCalls(t, "Save", 1)
	})

	t.Run("failure: returns 500 when usecase returns other error", func(t *testing.T) {
		mockUc := mocks.NewMockUsecase(t)
		e := echo.New()
		handlerInst := web.NewHandler(mockUc)

		expectedErr := errors.New("unique constraint violation")

		mockUc.EXPECT().
			Save(mock.Anything, mock.Anything).
			Return(expectedErr).Times(1)

		bodyBytes := []byte(`{"name":"Fail User","email":"fail@example.com","role":"user","permission_ids":[]}`)
		req := newRequest(http.MethodPost, "/api/v1/users", bodyBytes)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handlerErr := handlerInst.CreateUser(c)

		resp := extractResponse(t, handlerErr)

		assertFullResponse(t, resp, http.StatusInternalServerError, constants.ApplicationInternalServerError, constants.ApplicationErrorInternalServerError, "unique constraint violation", nil)
		mockUc.AssertNumberOfCalls(t, "Save", 1)
	})
}

// ===========================================================================
// GetUserByID — GET /api/v1/users/:id
// ===========================================================================
func TestHandler_GetUserByID(t *testing.T) {
	testID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	t.Run("success: returns user with all response fields", func(t *testing.T) {
		mockUc := mocks.NewMockUsecase(t)
		e := echo.New()
		handlerInst := web.NewHandler(mockUc)

		expectedUser := buildUser(testID, "findme@example.com", constants.RoleAdmin)

		mockUc.EXPECT().
			GetUserByID(mock.Anything, mock.Anything).
			Run(func(ctx context.Context, id uuid.UUID) {
				assert.NotNil(t, ctx)
				assert.Equal(t, testID, id)
			}).
			Return(expectedUser, nil).Times(1)

		req := newRequest(http.MethodGet, fmt.Sprintf("/api/v1/users/%s", testID))
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handlerErr := handlerInst.GetUserByID(c, testID)

		resp := extractResponse(t, handlerErr)

		// Success — core fields (Data is web.User from fn.CopyJSON with typed User field)
		assert.Equal(t, http.StatusOK, resp.HttpStatus)
		assert.Equal(t, constants.ApplicationCodeSuccess, resp.ResponseCode)
		assert.Empty(t, string(resp.ErrorType))
		assert.Equal(t, "successful", resp.Message)
	})

	t.Run("success: returns no-content when user not found", func(t *testing.T) {
		mockUc := mocks.NewMockUsecase(t)
		e := echo.New()
		handlerInst := web.NewHandler(mockUc)

		mockUc.EXPECT().
			GetUserByID(mock.Anything, mock.Anything).
			Return(nil, nil).Times(1)

		req := newRequest(http.MethodGet, fmt.Sprintf("/api/v1/users/%s", testID))
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handlerErr := handlerInst.GetUserByID(c, testID)

		resp := extractResponse(t, handlerErr)

		// NoContent response — ResponseCode is empty (not C20000)
		assert.Equal(t, http.StatusNoContent, resp.HttpStatus)
		assert.Empty(t, resp.ResponseCode)
		mockUc.AssertNumberOfCalls(t, "GetUserByID", 1)
	})

	t.Run("failure: returns 500 when usecase returns error", func(t *testing.T) {
		mockUc := mocks.NewMockUsecase(t)
		e := echo.New()
		handlerInst := web.NewHandler(mockUc)

		expectedErr := errors.New("query execution failed")

		mockUc.EXPECT().
			GetUserByID(mock.Anything, mock.Anything).
			Return(nil, expectedErr).Times(1)

		req := newRequest(http.MethodGet, fmt.Sprintf("/api/v1/users/%s", testID))
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handlerErr := handlerInst.GetUserByID(c, testID)

		resp := extractResponse(t, handlerErr)

		assertFullResponse(t, resp, http.StatusInternalServerError, constants.ApplicationInternalServerError, constants.ApplicationErrorInternalServerError, "query execution failed", nil)
		mockUc.AssertNumberOfCalls(t, "GetUserByID", 1)
	})
}

// ===========================================================================
// DeleteUserByID — DELETE /api/v1/users/:id
// ===========================================================================
func TestHandler_DeleteUserByID(t *testing.T) {
	t.Run("success: deletes user with all response fields", func(t *testing.T) {
		mockUc := mocks.NewMockUsecase(t)
		e := echo.New()
		handlerInst := web.NewHandler(mockUc)

		testID := uuid.MustParse("880e8400-e29b-41d4-a716-446655440000")

		mockUc.EXPECT().
			Delete(mock.Anything, mock.Anything).
			Run(func(ctx context.Context, id uuid.UUID) {
				assert.NotNil(t, ctx)
				assert.Equal(t, testID, id)
			}).
			Return(nil).Times(1)

		req := newRequest(http.MethodDelete, fmt.Sprintf("/api/v1/users/%s", testID))
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handlerErr := handlerInst.DeleteUserByID(c, testID)

		resp := extractResponse(t, handlerErr)

		assertFullResponse(t, resp, http.StatusOK, constants.ApplicationCodeSuccess, "", "successful", nil)
		mockUc.AssertNumberOfCalls(t, "Delete", 1)
	})

	t.Run("failure: returns 409 when user already removed", func(t *testing.T) {
		mockUc := mocks.NewMockUsecase(t)
		e := echo.New()
		handlerInst := web.NewHandler(mockUc)

		testID := uuid.MustParse("990e8400-e29b-41d4-a716-446655440000")

		conflictErr := errors.New(string(constants.ApplicationCodeConflictUserAlreadyRemoved))

		mockUc.EXPECT().
			Delete(mock.Anything, mock.Anything).
			Run(func(ctx context.Context, id uuid.UUID) {
				assert.NotNil(t, ctx)
				assert.Equal(t, testID, id)
			}).
			Return(conflictErr).Times(1)

		req := newRequest(http.MethodDelete, fmt.Sprintf("/api/v1/users/%s", testID))
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handlerErr := handlerInst.DeleteUserByID(c, testID)

		resp := extractResponse(t, handlerErr)

		assertFullResponse(t, resp, http.StatusConflict, constants.ApplicationCodeConflictUserAlreadyRemoved, constants.ApplicationErrorConflict, "user id already removed: "+testID.String(), nil)
		mockUc.AssertNumberOfCalls(t, "Delete", 1)
	})

	t.Run("failure: returns 500 when usecase returns other error", func(t *testing.T) {
		mockUc := mocks.NewMockUsecase(t)
		e := echo.New()
		handlerInst := web.NewHandler(mockUc)

		testID := uuid.MustParse("aa0e8400-e29b-41d4-a716-446655440000")

		expectedErr := errors.New("transaction rollback on commit")

		mockUc.EXPECT().
			Delete(mock.Anything, mock.Anything).
			Run(func(ctx context.Context, id uuid.UUID) {
				assert.NotNil(t, ctx)
				assert.Equal(t, testID, id)
			}).
			Return(expectedErr).Times(1)

		req := newRequest(http.MethodDelete, fmt.Sprintf("/api/v1/users/%s", testID))
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handlerErr := handlerInst.DeleteUserByID(c, testID)

		resp := extractResponse(t, handlerErr)

		assertFullResponse(t, resp, http.StatusInternalServerError, constants.ApplicationInternalServerError, constants.ApplicationErrorInternalServerError, "transaction rollback on commit", nil)
		mockUc.AssertNumberOfCalls(t, "Delete", 1)
	})
}

// ===========================================================================
// UpdateUserByID — PUT /api/v1/users/:id
// ===========================================================================
func TestHandler_UpdateUserByID(t *testing.T) {
	testID := uuid.MustParse("bb0e8400-e29b-41d4-a716-446655440000")

	t.Run("success: updates user and returns all response fields", func(t *testing.T) {
		mockUc := mocks.NewMockUsecase(t)
		e := echo.New()
		handlerInst := web.NewHandler(mockUc)

		existingUser := buildUser(testID, "old@example.com", constants.RoleUser)

		bodyBytes := []byte(`{"name":"Updated User","email":"updated@example.com","role":"admin","permission_ids":["aaaa1111-2222-3333-4444-555555555555","bbbb2222-3333-4444-5555-666666666666"]}`)
		mockUc.EXPECT().
			GetUserByID(mock.Anything, testID).
			Return(existingUser, nil).Times(1)
		mockUc.EXPECT().
			Save(mock.Anything, mock.Anything).
			Run(func(ctx context.Context, u *models.User) {
				assert.NotNil(t, ctx)
				assert.Equal(t, testID, u.ID)
				assert.Equal(t, "Updated User", u.Name.ValueOrZero())
				assert.Equal(t, "updated@example.com", u.Email.ValueOrZero())
			}).
			Return(nil).Times(1)

		req := newRequest(http.MethodPut, fmt.Sprintf("/api/v1/users/%s", testID), bodyBytes)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handlerErr := handlerInst.UpdateUserByID(c, testID)

		resp := extractResponse(t, handlerErr)

		assert.Equal(t, http.StatusOK, resp.HttpStatus)
		assert.Equal(t, constants.ApplicationCodeSuccess, resp.ResponseCode)
		assert.Empty(t, string(resp.ErrorType))
		assert.Equal(t, "successful", resp.Message)

		mockUc.AssertNumberOfCalls(t, "GetUserByID", 1)
		mockUc.AssertNumberOfCalls(t, "Save", 1)
	})

	t.Run("success: returns no-content when user not found", func(t *testing.T) {
		mockUc := mocks.NewMockUsecase(t)
		e := echo.New()
		handlerInst := web.NewHandler(mockUc)

		bodyBytes := []byte(`{"name":"Updated","email":"updated@example.com","role":"user","permission_ids":[]}`)
		mockUc.EXPECT().
			GetUserByID(mock.Anything, testID).
			Return(nil, nil).Times(1)

		req := newRequest(http.MethodPut, fmt.Sprintf("/api/v1/users/%s", testID), bodyBytes)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handlerErr := handlerInst.UpdateUserByID(c, testID)

		resp := extractResponse(t, handlerErr)

		assert.Equal(t, http.StatusNoContent, resp.HttpStatus)
		assert.Empty(t, resp.ResponseCode)
		mockUc.AssertNumberOfCalls(t, "GetUserByID", 1)
	})

	t.Run("failure: returns 409 when email is duplicated after update", func(t *testing.T) {
		mockUc := mocks.NewMockUsecase(t)
		e := echo.New()
		handlerInst := web.NewHandler(mockUc)

		existingUser := buildUser(testID, "original@example.com", constants.RoleUser)

		bodyBytes := []byte(`{"name":"Updated","email":"taken@example.com","role":"user","permission_ids":[]}`)
		conflictErr := errors.New(string(constants.ApplicationCodeConflictUserEmailWasDuplicate))

		mockUc.EXPECT().
			GetUserByID(mock.Anything, testID).
			Return(existingUser, nil).Times(1)
		mockUc.EXPECT().
			Save(mock.Anything, mock.Anything).
			Return(conflictErr).Times(1)

		req := newRequest(http.MethodPut, fmt.Sprintf("/api/v1/users/%s", testID), bodyBytes)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handlerErr := handlerInst.UpdateUserByID(c, testID)

		resp := extractResponse(t, handlerErr)

		assertFullResponse(t, resp, http.StatusConflict, constants.ApplicationCodeConflictUserEmailWasDuplicate, constants.ApplicationErrorConflict, "email taken@example.com was duplicate", nil)
	})

	t.Run("failure: returns 500 when usecase returns other save error", func(t *testing.T) {
		mockUc := mocks.NewMockUsecase(t)
		e := echo.New()
		handlerInst := web.NewHandler(mockUc)

		existingUser := buildUser(testID, "original@example.com", constants.RoleUser)

		bodyBytes := []byte(`{"name":"Updated","email":"updated@example.com","role":"user","permission_ids":[]}`)
		expectedErr := errors.New("unique constraint violation on email_idx")

		mockUc.EXPECT().
			GetUserByID(mock.Anything, testID).
			Return(existingUser, nil).Times(1)
		mockUc.EXPECT().
			Save(mock.Anything, mock.Anything).
			Return(expectedErr).Times(1)

		req := newRequest(http.MethodPut, fmt.Sprintf("/api/v1/users/%s", testID), bodyBytes)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handlerErr := handlerInst.UpdateUserByID(c, testID)

		resp := extractResponse(t, handlerErr)

		assertFullResponse(t, resp, http.StatusInternalServerError, constants.ApplicationInternalServerError, constants.ApplicationErrorInternalServerError, "unique constraint violation on email_idx", nil)
	})
}
