package server

import (
	"fmt"
	"net"
	"net/http"

	helperMiddl "git.innovasive.co.th/backend/helper/middleware"
	helperRoute "git.innovasive.co.th/backend/helper/route"
	"git.innovasive.co.th/backend/psql"
	_ "github.com/BlackMocca/go-clean-template/integration"
	myMiddL "github.com/BlackMocca/go-clean-template/middleware"
	"github.com/BlackMocca/go-clean-template/route"
	_user_handler "github.com/BlackMocca/go-clean-template/service/user/http"
	_user_repository "github.com/BlackMocca/go-clean-template/service/user/repository"
	_user_usecase "github.com/BlackMocca/go-clean-template/service/user/usecase"
	_user_validator "github.com/BlackMocca/go-clean-template/service/user/validator"
	"github.com/getsentry/sentry-go"
	sentryecho "github.com/getsentry/sentry-go/echo"
	"github.com/labstack/echo/v4"
	echoMiddL "github.com/labstack/echo/v4/middleware"
	"google.golang.org/grpc"
)

type Server struct {
	ServerReady  chan bool
	APP_LOGGER   bool
	APP_PORT     string
	JWT_SECRET   string
	GRPC_PORT    string
	GRPC_TIMEOUT int

	SENTRY_DSN string

	PsqlDB *psql.Client
}

func (s Server) Start() {
	sentryErr := sentry.Init(sentry.ClientOptions{
		Dsn: s.SENTRY_DSN,
	})
	/* init grpc */
	server := grpc.NewServer()
	defer server.GracefulStop()

	e := echo.New()
	e.HTTPErrorHandler = helperMiddl.SentryCapture(e)
	helperRoute.RegisterVersion(e)
	if s.APP_LOGGER {
		e.Use(echoMiddL.Logger())
	}
	e.Use(echoMiddL.Recover())
	e.Use(sentryecho.New(sentryecho.Options{Repanic: true}))

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	middL := myMiddL.InitMiddleware(s.JWT_SECRET)
	e.Use(echoMiddL.Recover())
	e.Use(echoMiddL.CORSWithConfig(echoMiddL.CORSConfig{
		Skipper:      echoMiddL.DefaultSkipper,
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete},
	}))
	e.Use(middL.InitContextIfNotExists)
	e.Use(middL.InputForm)

	/* Inject Repository */

	userRepo := _user_repository.NewPsqlUserRepository(s.PsqlDB)

	/* Inject Usecase */

	userUs := _user_usecase.NewUserUsecase(userRepo)

	/* Inject Handler */

	handler := _user_handler.NewUserHandler(e, userUs)

	/* validation */
	userValidation := _user_validator.Validation{}

	/* route */
	r := route.NewRoute(e, middL)
	r.RegisterRouteUser(handler, userValidation)

	/* serve gprc */
	go func() {
		if r := recover(); r != nil {
			fmt.Println(r.(error))
		}
		s.startGRPCServer(server)
	}()

	if s.ServerReady != nil {
		s.ServerReady <- true
	}

	/* serve echo */
	port := fmt.Sprintf(":%s", s.APP_PORT)
	if sentryErr == nil {
		sentry.CaptureException(e.Start(port))
	} else {
		e.Logger.Fatal(e.Start(port))
	}
}

func (s *Server) startGRPCServer(server *grpc.Server) {
	listen, err := net.Listen("tcp", fmt.Sprintf(":%s", s.GRPC_PORT))
	if err != nil {
		panic("failed to listen: " + err.Error())
	}

	/* serve grpc */
	fmt.Println(fmt.Sprintf("Start grpc Server [::%s]", s.GRPC_PORT))
	if err := server.Serve(listen); err != nil {
		panic(err)
	}
}
