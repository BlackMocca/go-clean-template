package web

import (
	"example/constants"
	"example/middleware"
	"example/models"
	"example/service/user"
	"fmt"
	"net/http"
	"strings"

	"github.com/BlackMocca/utils/fn"
	utilModels "github.com/BlackMocca/utils/models"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type handler struct {
	userUs user.Usecase
}

func NewHandler(userUs user.Usecase) ServerInterface {
	return &handler{userUs: userUs}
}

// ตรวจสอบสถานะของ API service
// (GET /api/health)
func (h handler) HealthCheck(ctx echo.Context) error {
	return middleware.NewResponseJSON(nil)
}

// ดึงรายการผู้ใช้ทั้งหมด
// (GET /api/v1/users)
func (h handler) ListUsers(ctx echo.Context, params ListUsersParams) error {
	args, err := fn.ConvertStruct[ListUsersParams, user.ListParams](params)
	if err != nil {
		return middleware.NewResponseError(http.StatusInternalServerError, constants.ApplicationInternalServerError, constants.ApplicationErrorInvalidArgument, err.Error(), params)
	}

	paginator, err := h.userUs.ListUsers(ctx.Request().Context(), args)
	if err != nil {
		return middleware.NewResponseError(http.StatusInternalServerError, constants.ApplicationInternalServerError, constants.ApplicationErrorInternalServerError, err.Error(), nil)
	}
	if len(paginator.Documents) == 0 {
		return middleware.NewResponseNoContent()
	}

	resp := ListUsersSuccess{}
	if err := fn.CopyJSON(paginator, &resp.Data); err != nil {
		return middleware.NewResponseError(http.StatusInternalServerError, constants.ApplicationInternalServerError, constants.ApplicationErrorInternalServerError, err.Error(), nil)
	}
	return middleware.NewResponseJSON(resp.Data)
}

// สร้างผู้ใช้ใหม่
// (POST /api/v1/users)
func (h handler) CreateUser(ctx echo.Context) error {
	var rawctx = ctx.Request().Context()
	var reqBody = UpdateUserByIDJSONRequestBody{}
	if err := ctx.Bind(&reqBody); err != nil {
		return middleware.NewResponseError(http.StatusInternalServerError, constants.ApplicationCodeBadRequestInvalidArgument, constants.ApplicationErrorInvalidArgument, err.Error(), nil)
	}

	var user = new(models.User)
	user.ID = uuid.New()
	if err := fn.CopyJSON(reqBody, user); err != nil {
		return middleware.NewResponseError(http.StatusInternalServerError, constants.ApplicationInternalServerError, constants.ApplicationErrorInvalidArgument, err.Error(), nil)
	}
	var permissions = make([]*models.Permission, 0, len(reqBody.PermissionIds))
	for _, permisId := range reqBody.PermissionIds {
		permissions = append(permissions, &models.Permission{Id: uuid.MustParse(permisId.String())})
	}
	user.Permissions.Set(permissions)
	user.CreatedAt = utilModels.NewTimestampFromNow().ToPointer()
	user.UpdatedAt = utilModels.NewTimestampFromNow().ToPointer()

	if err := h.userUs.Save(rawctx, user); err != nil {
		if strings.EqualFold(err.Error(), string(constants.ApplicationCodeConflictUserEmailWasDuplicate)) {
			return middleware.NewResponseError(http.StatusConflict, constants.ApplicationCodeConflictUserEmailWasDuplicate, constants.ApplicationErrorConflict, fmt.Sprintf("email %s was duplicated", user.Email.ValueOrZero()), nil)
		}
		return middleware.NewResponseError(http.StatusInternalServerError, constants.ApplicationInternalServerError, constants.ApplicationErrorInternalServerError, err.Error(), nil)
	}

	resp := GetUserByIDSuccess{}
	fn.CopyJSON(user, &resp.Data)
	return middleware.NewResponseJSON(resp.Data)
}

// ลบผู้ใช้
// (DELETE /api/v1/users/{id})
func (h handler) DeleteUserByID(ctx echo.Context, id DocumentID) error {
	var rawctx = ctx.Request().Context()

	if err := h.userUs.Delete(rawctx, id); err != nil {
		if strings.EqualFold(err.Error(), string(constants.ApplicationCodeConflictUserAlreadyRemoved)) {
			return middleware.NewResponseError(http.StatusConflict, constants.ApplicationCodeConflictUserAlreadyRemoved, constants.ApplicationErrorConflict, fmt.Sprintf("user id already removed: %s", id), nil)
		}
		return middleware.NewResponseError(http.StatusInternalServerError, constants.ApplicationInternalServerError, constants.ApplicationErrorInternalServerError, err.Error(), nil)
	}

	return middleware.NewResponseJSON(nil)
}

// ดึงข้อมูลผู้ใช้ตาม ID
// (GET /api/v1/users/{id})
func (h handler) GetUserByID(ctx echo.Context, id DocumentID) error {
	user, err := h.userUs.GetUserByID(ctx.Request().Context(), id)
	if err != nil {
		return middleware.NewResponseError(http.StatusInternalServerError, constants.ApplicationInternalServerError, constants.ApplicationErrorInternalServerError, err.Error(), nil)
	}
	if user == nil {
		return middleware.NewResponseNoContent()
	}

	resp := GetUserByIDSuccess{}
	fn.CopyJSON(user, &resp.Data)
	return middleware.NewResponseJSON(resp.Data)
}

// อัพเดทข้อมูลผู้ใช้
// (PUT /api/v1/users/{id})
func (h handler) UpdateUserByID(ctx echo.Context, id DocumentID) error {
	var rawctx = ctx.Request().Context()
	var reqBody = UpdateUserByIDJSONRequestBody{}
	if err := ctx.Bind(&reqBody); err != nil {
		return middleware.NewResponseError(http.StatusInternalServerError, constants.ApplicationCodeBadRequestInvalidArgument, constants.ApplicationErrorInvalidArgument, err.Error(), nil)
	}

	user, err := h.userUs.GetUserByID(ctx.Request().Context(), id)
	if err != nil {
		return middleware.NewResponseError(http.StatusInternalServerError, constants.ApplicationInternalServerError, constants.ApplicationErrorInternalServerError, err.Error(), nil)
	}
	if user == nil {
		return middleware.NewResponseNoContent()
	}

	if err := fn.CopyJSON(reqBody, user); err != nil {
		return middleware.NewResponseError(http.StatusInternalServerError, constants.ApplicationInternalServerError, constants.ApplicationErrorInvalidArgument, err.Error(), nil)
	}
	var permissions = make([]*models.Permission, 0, len(reqBody.PermissionIds))
	for _, permisId := range reqBody.PermissionIds {
		permissions = append(permissions, &models.Permission{Id: uuid.MustParse(permisId.String())})
	}
	user.Permissions.Set(permissions)
	user.UpdatedAt = utilModels.NewTimestampFromNow().ToPointer()

	if err := h.userUs.Save(rawctx, user); err != nil {
		if strings.EqualFold(err.Error(), string(constants.ApplicationCodeConflictUserEmailWasDuplicate)) {
			return middleware.NewResponseError(http.StatusConflict, constants.ApplicationCodeConflictUserEmailWasDuplicate, constants.ApplicationErrorConflict, fmt.Sprintf("email %s was duplicate", user.Email.ValueOrZero()), nil)
		}
		return middleware.NewResponseError(http.StatusInternalServerError, constants.ApplicationInternalServerError, constants.ApplicationErrorInternalServerError, err.Error(), nil)
	}

	resp := GetUserByIDSuccess{}
	fn.CopyJSON(user, &resp.Data)
	return middleware.NewResponseJSON(resp.Data)
}
