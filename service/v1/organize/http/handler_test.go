package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	helperModel "git.innovasive.co.th/backend/models"
	"github.com/Blackmocca/go-clean-template/constants"
	"github.com/Blackmocca/go-clean-template/middleware"
	_middleware_mock "github.com/Blackmocca/go-clean-template/middleware/mocks"
	"github.com/Blackmocca/go-clean-template/models"
	"github.com/Blackmocca/go-clean-template/route"
	_organize_mock "github.com/Blackmocca/go-clean-template/service/v1/organize/mocks"
	"github.com/Blackmocca/go-clean-template/service/v1/organize/validator"
	"github.com/gofrs/uuid"
	"github.com/guregu/null/zero"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var middl = middleware.InitMiddleware("")

var jwtPayloadStr = `{
	"id": "e9273fb9-4028-48d4-9975-82f18763f71d",
	"username": "yongut@gmail.com"
}`

func getMiddleware() middleware.GoMiddlewareInf {
	middlInf := new(_middleware_mock.GoMiddlewareInf)

	var isAuth = func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {

			return next(c)
		}
	}

	var setpayload = func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			var jwtPayload = map[string]interface{}{}
			json.Unmarshal([]byte(jwtPayloadStr), &jwtPayload)

			c.Set("payload", jwtPayload)

			return next(c)
		}
	}

	middlInf.On("IsAuthorization", mock.AnythingOfType("echo.HandlerFunc")).Return(isAuth).Maybe()
	middlInf.On("SetPayload", mock.AnythingOfType("echo.HandlerFunc")).Return(setpayload).Maybe()

	/* example of another middleware */
	middlInf.On("InitContextIfNotExists", mock.AnythingOfType("echo.HandlerFunc")).Return(middl.InitContextIfNotExists).Maybe()
	middlInf.On("InputForm", mock.AnythingOfType("echo.HandlerFunc")).Return(middl.InputForm).Maybe()
	middlInf.On("ValidateParamId", mock.AnythingOfType("string")).Return(middl.ValidateParamId).Maybe()
	middlInf.On("RequireQueryParam", mock.AnythingOfType("string")).Return(middl.RequireQueryParam).Maybe()
	middlInf.On("ValidateRequiredHeader", mock.AnythingOfType("string")).Return(middl.ValidateRequiredHeader).Maybe()
	return middlInf
}

func TestFetchAll_Success(t *testing.T) {
	now := helperModel.NewTimestampFromTime(time.Now())
	orgId1 := uuid.FromStringOrNil("907eefd8-181b-457b-8ca2-692c442b2b0b")
	orgId2 := uuid.FromStringOrNil("97478e1b-2ebd-4dee-88da-49da3ca482f4")
	orgId3 := uuid.FromStringOrNil("1f66d3c9-a549-46de-9639-0f6ff6c8d7f3")
	orgs := []*models.Organize{
		&models.Organize{
			Id:        &orgId1,
			Name:      "จราจรการสื่อสาร",
			AliasName: zero.StringFrom("จส100"),
			OrgType:   constants.ORGANIZE_TYPE_PUBLIC,
			CreatedAt: &now,
			UpdatedAt: &now,
		},
		&models.Organize{
			Id:        &orgId2,
			Name:      "ตุ้ดซี่ review",
			AliasName: zero.StringFrom("ตซ"),
			OrgType:   constants.ORGANIZE_TYPE_PUBLIC,
			CreatedAt: &now,
			UpdatedAt: &now,
		},
		&models.Organize{
			Id:        &orgId3,
			Name:      "เกมถูกบอกด้วย",
			AliasName: zero.StringFrom("เกมถูกบอกด้วย"),
			OrgType:   constants.ORGANIZE_TYPE_PUBLIC,
			CreatedAt: &now,
			UpdatedAt: &now,
		},
	}

	organizeUs := new(_organize_mock.OrganizeUsecase)
	organizeUs.On("FetchAll", mock.Anything, mock.AnythingOfType("*sync.Map"), mock.AnythingOfType("*models.Paginator")).Return(orgs, nil).Once().Run(func(args mock.Arguments) {
		ctx := args.Get(0)
		m := args.Get(1).(*sync.Map)
		p := args.Get(2)

		assert.NotNil(t, ctx)
		assert.NotNil(t, m)
		assert.Nil(t, p)
	})

	e := echo.New()
	e.Use(middl.InitContextIfNotExists)
	e.Use(middl.InputForm)
	e.Use(middl.SetTracer)
	req := httptest.NewRequest(http.MethodGet, "/v1/list/all", nil)
	rec := httptest.NewRecorder()

	handler := NewOrganizeHandler(organizeUs)
	middl := getMiddleware()
	api := route.NewRoute(e, middl)
	validate := validator.Validation{}
	api.RegisterOrganization(handler, validate)

	e.ServeHTTP(rec, req)

	var result map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &result)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.NotEmpty(t, result["organizes"])
}
func TestFetchOneById_Success(t *testing.T) {
	now := helperModel.NewTimestampFromTime(time.Now())
	orgId1 := uuid.FromStringOrNil("042e37e5-3027-4499-9a02-91ade81f2d67")

	org := &models.Organize{
		Id:        &orgId1,
		Name:      "เกมถูกบอกด้วย",
		AliasName: zero.StringFrom("เกมถูกบอกด้วย"),
		OrgType:   constants.ORGANIZE_TYPE_PUBLIC,
		CreatedAt: &now,
		UpdatedAt: &now,
	}

	organizeUs := new(_organize_mock.OrganizeUsecase)
	organizeUs.On("FetchOneById", mock.Anything, mock.AnythingOfType("*uuid.UUID")).Return(org, nil).Once().Run(func(args mock.Arguments) {
		ctx := args.Get(0)
		m := args.Get(1).(*uuid.UUID)

		assert.NotNil(t, ctx)
		assert.NotNil(t, m)
	})

	e := echo.New()
	e.Use(middl.InitContextIfNotExists)
	e.Use(middl.InputForm)
	e.Use(middl.SetTracer)
	req := httptest.NewRequest(http.MethodGet, "/v1/list/042e37e5-3027-4499-9a02-91ade81f2d67", nil)
	rec := httptest.NewRecorder()

	handler := NewOrganizeHandler(organizeUs)
	middl := getMiddleware()
	api := route.NewRoute(e, middl)
	validate := validator.Validation{}
	api.RegisterOrganization(handler, validate)

	e.ServeHTTP(rec, req)

	var result map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &result)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.NotEmpty(t, result["organizes"])
}

func TestFetchOneById_Error500(t *testing.T) {

	organizeUs := new(_organize_mock.OrganizeUsecase)
	organizeUs.On("FetchOneById", mock.Anything, mock.AnythingOfType("*uuid.UUID")).Return(nil, errors.New("unexpected")).Once().Run(func(args mock.Arguments) {
		ctx := args.Get(0)
		m := args.Get(1).(*uuid.UUID)

		assert.NotNil(t, ctx)
		assert.NotNil(t, m)
	})

	e := echo.New()
	e.Use(middl.InitContextIfNotExists)
	e.Use(middl.InputForm)
	e.Use(middl.SetTracer)
	req := httptest.NewRequest(http.MethodGet, "/v1/list/042e37e5-3027-4499-9a02-91ade81f2d67", nil)
	rec := httptest.NewRecorder()

	handler := NewOrganizeHandler(organizeUs)
	middl := getMiddleware()
	api := route.NewRoute(e, middl)
	validate := validator.Validation{}
	api.RegisterOrganization(handler, validate)

	e.ServeHTTP(rec, req)

	var result map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &result)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestCreateOrg_Success(t *testing.T) {
	body := map[string]interface{}{
		"name":           "หน่วยทำลายวเครื่องดื่มนิลาแบบจู่โจม",
		"alias_name":     "ทวจ.",
		"org_type":       "PUBLIC",
		"private_tel_no": "112341",
		"admin_1":        "1f66d3c9-a549-46de-9639-0f6ff6c8d7f3",
		"admin_2":        "1f66d3c9-a549-46de-9639-0f6ff6c8d7f3",
	}

	payload, _ := json.Marshal(body)

	organizeUs := new(_organize_mock.OrganizeUsecase)
	organizeUs.On("Create", mock.Anything, mock.AnythingOfType("*models.Organize")).Return(nil).Once().Run(func(args mock.Arguments) {
		ctx := args.Get(0)
		organize := args.Get(1).(*models.Organize)

		assert.NotNil(t, ctx)
		assert.Equal(t, organize.Name, body["name"])
		assert.Equal(t, organize.AliasName.ValueOrZero(), body["alias_name"])
		assert.Equal(t, organize.OrgType, body["org_type"])
		assert.Equal(t, organize.Admin1.String(), body["admin_1"])
		assert.Equal(t, organize.Admin2.String(), body["admin_2"])
	})

	e := echo.New()
	e.Use(middl.InitContextIfNotExists)
	e.Use(middl.InputForm)
	e.Use(middl.SetTracer)
	req := httptest.NewRequest(http.MethodPost, "/v1/organizes", strings.NewReader(string(payload)))
	req.Header.Add("Content-Type", echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	handler := NewOrganizeHandler(organizeUs)
	middl := getMiddleware()
	api := route.NewRoute(e, middl)
	validate := validator.Validation{}
	api.RegisterOrganization(handler, validate)

	e.ServeHTTP(rec, req)

	var result map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &result)

	assert.Equal(t, http.StatusOK, rec.Code)
}
func TestCreateOrg_Error500(t *testing.T) {
	body := map[string]interface{}{
		"name":           "หน่วยทำลายวเครื่องดื่มนิลาแบบจู่โจม",
		"alias_name":     "ทวจ.",
		"org_type":       "PUBLIC",
		"private_tel_no": "112341",
		"admin_1":        "1f66d3c9-a549-46de-9639-0f6ff6c8d7f3",
		"admin_2":        "1f66d3c9-a549-46de-9639-0f6ff6c8d7f3",
	}

	payload, _ := json.Marshal(body)

	organizeUs := new(_organize_mock.OrganizeUsecase)
	organizeUs.On("Create", mock.Anything, mock.AnythingOfType("*models.Organize")).Return(errors.New("unexpected")).Once().Run(func(args mock.Arguments) {
		ctx := args.Get(0)
		organize := args.Get(1).(*models.Organize)

		assert.NotNil(t, ctx)
		assert.Equal(t, organize.Name, body["name"])
		assert.Equal(t, organize.AliasName.ValueOrZero(), body["alias_name"])
		assert.Equal(t, organize.OrgType, body["org_type"])
		assert.Equal(t, organize.Admin1.String(), body["admin_1"])
		assert.Equal(t, organize.Admin2.String(), body["admin_2"])
	})

	e := echo.New()
	e.Use(middl.InitContextIfNotExists)
	e.Use(middl.InputForm)
	e.Use(middl.SetTracer)
	req := httptest.NewRequest(http.MethodPost, "/v1/organizes", strings.NewReader(string(payload)))
	req.Header.Add("Content-Type", echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	handler := NewOrganizeHandler(organizeUs)
	middl := getMiddleware()
	api := route.NewRoute(e, middl)
	validate := validator.Validation{}
	api.RegisterOrganization(handler, validate)

	e.ServeHTTP(rec, req)

	var result map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &result)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestCreateOrg_ErrorNameWasDuplicate(t *testing.T) {
	body := map[string]interface{}{
		"name":           "หน่วยทำลายวเครื่องดื่มนิลาแบบจู่โจม",
		"alias_name":     "ทวจ.",
		"org_type":       "PUBLIC",
		"private_tel_no": "112341",
		"admin_1":        "1f66d3c9-a549-46de-9639-0f6ff6c8d7f3",
		"admin_2":        "1f66d3c9-a549-46de-9639-0f6ff6c8d7f4",
	}

	payload, _ := json.Marshal(body)

	organizeUs := new(_organize_mock.OrganizeUsecase)
	organizeUs.On("Create", mock.Anything, mock.AnythingOfType("*models.Organize")).Return(errors.New(constants.ERROR_ORGANIZE_NAME_WAS_DUPLICATE)).Once().Run(func(args mock.Arguments) {
		ctx := args.Get(0)
		organize := args.Get(1).(*models.Organize)

		assert.NotNil(t, ctx)
		assert.Equal(t, organize.Name, body["name"])
		assert.Equal(t, organize.AliasName.ValueOrZero(), body["alias_name"])
		assert.Equal(t, organize.OrgType, body["org_type"])
		assert.Equal(t, organize.Admin1.String(), body["admin_1"])
		assert.Equal(t, organize.Admin2.String(), body["admin_2"])
	})

	e := echo.New()
	e.Use(middl.InitContextIfNotExists)
	e.Use(middl.InputForm)
	e.Use(middl.SetTracer)
	req := httptest.NewRequest(http.MethodPost, "/v1/organizes", strings.NewReader(string(payload)))
	req.Header.Add("Content-Type", echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	handler := NewOrganizeHandler(organizeUs)
	middl := getMiddleware()
	api := route.NewRoute(e, middl)
	validate := validator.Validation{}
	api.RegisterOrganization(handler, validate)

	e.ServeHTTP(rec, req)

	var result map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &result)

	assert.Equal(t, http.StatusConflict, rec.Code)
}

func TestCreateOrg_ErrorAliasNameWasDuplicate(t *testing.T) {
	body := map[string]interface{}{
		"name":           "หน่วยทำลายวเครื่องดื่มนิลาแบบจู่โจม",
		"alias_name":     "ทวจ.",
		"org_type":       "PUBLIC",
		"private_tel_no": "112341",
		"admin_1":        "1f66d3c9-a549-46de-9639-0f6ff6c8d7f3",
		"admin_2":        "1f66d3c9-a549-46de-9639-0f6ff6c8d7f4",
	}

	payload, _ := json.Marshal(body)

	organizeUs := new(_organize_mock.OrganizeUsecase)
	organizeUs.On("Create", mock.Anything, mock.AnythingOfType("*models.Organize")).Return(errors.New(constants.ERROR_ORGANIZE_ALIAS_NAME_WAS_DUPLICATE)).Once().Run(func(args mock.Arguments) {
		ctx := args.Get(0)
		organize := args.Get(1).(*models.Organize)

		assert.NotNil(t, ctx)
		assert.Equal(t, organize.Name, body["name"])
		assert.Equal(t, organize.AliasName.ValueOrZero(), body["alias_name"])
		assert.Equal(t, organize.OrgType, body["org_type"])
		assert.Equal(t, organize.Admin1.String(), body["admin_1"])
		assert.Equal(t, organize.Admin2.String(), body["admin_2"])
	})

	e := echo.New()
	e.Use(middl.InitContextIfNotExists)
	e.Use(middl.InputForm)
	e.Use(middl.SetTracer)
	req := httptest.NewRequest(http.MethodPost, "/v1/organizes", strings.NewReader(string(payload)))
	req.Header.Add("Content-Type", echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	handler := NewOrganizeHandler(organizeUs)
	middl := getMiddleware()
	api := route.NewRoute(e, middl)
	validate := validator.Validation{}
	api.RegisterOrganization(handler, validate)

	e.ServeHTTP(rec, req)

	var result map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &result)

	assert.Equal(t, http.StatusConflict, rec.Code)
}
func TestCreateOrg_ErrorPrivateTelNoWasDuplicate(t *testing.T) {
	body := map[string]interface{}{
		"name":           "หน่วยทำลายวเครื่องดื่มนิลาแบบจู่โจม",
		"alias_name":     "ทวจ.",
		"org_type":       "PUBLIC",
		"private_tel_no": "112341",
		"admin_1":        "1f66d3c9-a549-46de-9639-0f6ff6c8d7f3",
		"admin_2":        "1f66d3c9-a549-46de-9639-0f6ff6c8d7f4",
	}

	payload, _ := json.Marshal(body)

	organizeUs := new(_organize_mock.OrganizeUsecase)
	organizeUs.On("Create", mock.Anything, mock.AnythingOfType("*models.Organize")).Return(errors.New(constants.ERROR_ORGANIZE_PRIVATE_TEL_NO_WAS_DUPLICATE)).Once().Run(func(args mock.Arguments) {
		ctx := args.Get(0)
		organize := args.Get(1).(*models.Organize)

		assert.NotNil(t, ctx)
		assert.Equal(t, organize.Name, body["name"])
		assert.Equal(t, organize.AliasName.ValueOrZero(), body["alias_name"])
		assert.Equal(t, organize.OrgType, body["org_type"])
		assert.Equal(t, organize.Admin1.String(), body["admin_1"])
		assert.Equal(t, organize.Admin2.String(), body["admin_2"])
	})

	e := echo.New()
	e.Use(middl.InitContextIfNotExists)
	e.Use(middl.InputForm)
	e.Use(middl.SetTracer)
	req := httptest.NewRequest(http.MethodPost, "/v1/organizes", strings.NewReader(string(payload)))
	req.Header.Add("Content-Type", echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	handler := NewOrganizeHandler(organizeUs)
	middl := getMiddleware()
	api := route.NewRoute(e, middl)
	validate := validator.Validation{}
	api.RegisterOrganization(handler, validate)

	e.ServeHTTP(rec, req)

	var result map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &result)

	assert.Equal(t, http.StatusConflict, rec.Code)
}
func TestFetchAll_NO_CONTENT(t *testing.T) {
	orgs := []*models.Organize{}

	organizeUs := new(_organize_mock.OrganizeUsecase)
	organizeUs.On("FetchAll", mock.Anything, mock.AnythingOfType("*sync.Map"), mock.AnythingOfType("*models.Paginator")).Return(orgs, nil).Once().Run(func(args mock.Arguments) {
		ctx := args.Get(0)
		m := args.Get(1).(*sync.Map)
		p := args.Get(2)

		assert.NotNil(t, ctx)
		assert.NotNil(t, m)
		assert.Nil(t, p)
	})

	e := echo.New()
	e.Use(middl.InitContextIfNotExists)
	e.Use(middl.InputForm)
	e.Use(middl.SetTracer)
	req := httptest.NewRequest(http.MethodGet, "/v1/list/all", nil)
	rec := httptest.NewRecorder()

	handler := NewOrganizeHandler(organizeUs)
	middl := getMiddleware()
	api := route.NewRoute(e, middl)
	validate := validator.Validation{}
	api.RegisterOrganization(handler, validate)

	e.ServeHTTP(rec, req)

	var result map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &result)

	assert.Equal(t, http.StatusNoContent, rec.Code)
}
