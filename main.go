package main

import (
	"log"
	"net/http"

	_conf "github.com/BlackMocca/go-clean-template/config"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	myMiddL "github.com/BlackMocca/go-clean-template/middleware"
	_user_handler "github.com/BlackMocca/go-clean-template/service/user/http"
	_user_repository "github.com/BlackMocca/go-clean-template/service/user/repository"
	_user_usecase "github.com/BlackMocca/go-clean-template/service/user/usecase"
	"github.com/jmoiron/sqlx"
)

func sqlDB() *sqlx.DB {
	var connstr = _conf.GetEnv("PSQL_DATABASE_URL", "postgres://postgres:postgres@psql_db:5432/app_example?sslmode=disable")
	db, err := _conf.NewPsqlConnection(connstr)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func main() {
	psqlDB := sqlDB()

	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	middL := myMiddL.InitMiddleware()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	/* Inject Repository */

	userRepo := _user_repository.NewPsqlUserRepository(psqlDB)

	/* Inject Usecase */

	userUs := _user_usecase.NewUserUsecase(userRepo)

	/* Inject Handler */

	_user_handler.NewUserHandler(e, middL, userUs)

	port := ":" + _conf.GetEnv("PORT", "3000")
	e.Logger.Fatal(e.Start(port))
}
