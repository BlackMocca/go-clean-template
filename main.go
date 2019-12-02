package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_conf "gitlab.com/km/go-kafka-playground/config"
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

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	port := ":" + Config.GetEnv("PORT", "3000")
	e.Logger.Fatal(e.Start(port))
}
