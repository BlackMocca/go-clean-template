package main

import (
	"net/http"

	"git.innovasive.co.th/backend/helper"
	"git.innovasive.co.th/backend/psql"
	"github.com/getsentry/sentry-go"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/spf13/cast"

	myMiddL "github.com/BlackMocca/go-clean-template/middleware"
	"github.com/BlackMocca/go-clean-template/route"
	_user_handler "github.com/BlackMocca/go-clean-template/service/user/http"
	_user_repository "github.com/BlackMocca/go-clean-template/service/user/repository"
	_user_usecase "github.com/BlackMocca/go-clean-template/service/user/usecase"
	_user_validator "github.com/BlackMocca/go-clean-template/service/user/validator"
	sentryecho "github.com/getsentry/sentry-go/echo"
	echoMiddL "github.com/labstack/echo/v4/middleware"
)

var (
	APP_LOGGER   = cast.ToBool(helper.GetENV("APP_LOGGER", "true"))
	APP_PORT     = helper.GetENV("APP_PORT", "3000")
	JWT_SECRET   = helper.GetENV("JWT_SECRET", "test")
	GRPC_PORT    = helper.GetENV("GRPC_PORT", "3100")
	GRPC_TIMEOUT = cast.ToInt(helper.GetENV("GRPC_TIMEOUT", "120"))

	SENTRY_DSN = helper.GetENV("SENTRY_DSN", "")

	PSQL_DATABASE_URL = helper.GetENV("PSQL_DATABASE_URL", "postgres://postgres:postgres@psql_db:5432/app_example?sslmode=disable")
)

func sqlDB(con string) *psql.Client {
	db, err := psql.NewPsqlConnection(con)
	if err != nil {
		panic(err)
	}
	return db
}

func main() {
	psqlDB := sqlDB(PSQL_DATABASE_URL)

	sentryErr := sentry.Init(sentry.ClientOptions{
		Dsn: SENTRY_DSN,
	})

	e := echo.New()
	if APP_LOGGER {
		e.Use(echoMiddL.Logger())
	}
	e.Use(echoMiddL.Recover())
	e.Use(sentryecho.New(sentryecho.Options{Repanic: true}))

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	middL := myMiddL.InitMiddleware(JWT_SECRET)
	e.Use(middleware.Recover())
	e.Use(echoMiddL.CORSWithConfig(echoMiddL.CORSConfig{
		Skipper:      echoMiddL.DefaultSkipper,
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete},
	}))
	e.Use(middL.InitContextIfNotExists)
	e.Use(middL.InputForm)

	/* Inject Repository */

	userRepo := _user_repository.NewPsqlUserRepository(psqlDB)

	/* Inject Usecase */

	userUs := _user_usecase.NewUserUsecase(userRepo)

	/* Inject Handler */

	handler := _user_handler.NewUserHandler(e, userUs)

	/* validation */
	userValidation := _user_validator.Validation{}

	/* route */
	r := route.NewRoute(e, middL)
	r.RegisterRouteUser(handler, userValidation)

	port := ":" + APP_PORT
	if sentryErr == nil {
		sentry.CaptureException(e.Start(port))
	} else {
		e.Logger.Fatal(e.Start(port))
	}
}
