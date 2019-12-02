package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_conf "gitlab.com/km/go-kafka-playground/config"

	myMiddL "gitlab.com/km/go-kafka-playground/middleware"
	_user_handler "gitlab.com/km/go-kafka-playground/service/user/http"
	_user_repository "gitlab.com/km/go-kafka-playground/service/user/repository"
	_user_usecase "gitlab.com/km/go-kafka-playground/service/user/usecase"
)

var (
	Config *_conf.Config
)

func init() {
	Config = _conf.NewConfig()
}

func main() {
	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	middL := myMiddL.InitMiddleware()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	/* Inject Repository */

	userRepo := _user_repository.NewPsqlUserRepository(Config.PGORM)

	/* Inject Usecase */

	userUs := _user_usecase.NewUserUsecase(userRepo)

	/* Inject Handler */

	_user_handler.NewUserHandler(e, middL, userUs)

	port := ":" + Config.GetEnv("PORT", "3000")
	e.Logger.Fatal(e.Start(port))
}
