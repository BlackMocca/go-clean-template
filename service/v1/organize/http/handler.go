package http

import (
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	helperModel "git.innovasive.co.th/backend/models"
	"github.com/Blackmocca/go-clean-template/constants"
	"github.com/Blackmocca/go-clean-template/models"
	"github.com/Blackmocca/go-clean-template/service/v1/organize"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
)

type organizeHandler struct {
	orgUs organize.OrganizeUsecase
}

func NewOrganizeHandler(orgUs organize.OrganizeUsecase) organize.OrganizeHandler {
	return &organizeHandler{
		orgUs: orgUs,
	}
}

func (o organizeHandler) FetchAll(c echo.Context) error {
	var ctx = c.Request().Context()
	var showDisabled, _ = strconv.ParseBool(c.QueryParam("show_disabled"))
	var searchword = c.QueryParam("search_word")
	var args = new(sync.Map)

	if searchword != "" {
		args.Store("search_word", searchword)
	}

	args.Store("show_disabled", showDisabled)
	orgs, err := o.orgUs.FetchAll(ctx, args, nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if len(orgs) == 0 {
		return echo.NewHTTPError(http.StatusNoContent)
	}

	resp := map[string]interface{}{
		"organizes": orgs,
	}
	return c.JSON(http.StatusOK, resp)
}

func (o organizeHandler) FetchOneById(c echo.Context) error {
	var ctx = c.Request().Context()
	var orgId = uuid.FromStringOrNil(c.Param("org_id"))

	orgs, err := o.orgUs.FetchOneById(ctx, &orgId)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	resp := map[string]interface{}{
		"organizes": orgs,
	}
	return c.JSON(http.StatusOK, resp)
}

func (o organizeHandler) CreateOrg(c echo.Context) error {
	var ctx = c.Request().Context()
	var params = c.Get("params").(map[string]interface{})
	var ti = helperModel.NewTimestampFromTime(time.Now())

	org := models.NewOrganizeWithParams(params, nil)
	org.NewUUID()
	org.SetCreatedAt(ti)
	org.SetUpdatedAt(ti)

	if err := o.orgUs.Create(ctx, org); err != nil {
		if strings.Contains(err.Error(), constants.ERROR_ORGANIZE_NAME_WAS_DUPLICATE) {
			return echo.NewHTTPError(http.StatusConflict, err.Error())
		}
		if strings.Contains(err.Error(), constants.ERROR_ORGANIZE_ALIAS_NAME_WAS_DUPLICATE) {
			return echo.NewHTTPError(http.StatusConflict, err.Error())
		}
		if strings.Contains(err.Error(), constants.ERROR_ORGANIZE_PRIVATE_TEL_NO_WAS_DUPLICATE) {
			return echo.NewHTTPError(http.StatusConflict, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	resp := map[string]interface{}{
		"message":     "Created",
		"organize_id": org.Id,
	}
	return c.JSON(http.StatusOK, resp)
}
