package main

import (
	"fmt"
	"net"
	"net/http"

	"git.innovasive.co.th/backend/helper"
	helperMiddl "git.innovasive.co.th/backend/helper/middleware"
	helperRoute "git.innovasive.co.th/backend/helper/route"
	"git.innovasive.co.th/backend/psql"
	myMiddL "github.com/Blackmocca/go-clean-template/middleware"
	route "github.com/Blackmocca/go-clean-template/route"
	_organize_grpc "github.com/Blackmocca/go-clean-template/service/v1/organize/grpc"
	_organize_handler "github.com/Blackmocca/go-clean-template/service/v1/organize/http"
	_organize_repository "github.com/Blackmocca/go-clean-template/service/v1/organize/repository"
	_organize_usecase "github.com/Blackmocca/go-clean-template/service/v1/organize/usecase"
	_organize_validator "github.com/Blackmocca/go-clean-template/service/v1/organize/validator"
	_util_tracing "github.com/Blackmocca/go-clean-template/utils/opentracing"
	"github.com/getsentry/sentry-go"
	sentryecho "github.com/getsentry/sentry-go/echo"
	"github.com/labstack/echo/v4"
	echoMiddL "github.com/labstack/echo/v4/middleware"
	otgrpc "github.com/opentracing-contrib/go-grpc"
	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
)

var (
	APP_PORT          = helper.GetENV("APP_PORT", "3000")
	GRPC_PORT         = helper.GetENV("GRPC_PORT", "3100")
	JWT_SECRET        = helper.GetENV("JWT_SECRET", "cleantamplate")
	PSQL_DATABASE_URL = helper.GetENV("PSQL_DATABASE_URL", "postgres://postgres:postgres@psql_db:5432/app_example?sslmode=disable")

	SENTRY_DSN = helper.GetENV("SENTRY_DSN", "")
)

func sqlDBWithTracing(con string, tracer opentracing.Tracer) *psql.Client {
	db, err := psql.NewPsqlWithTracingConnection(con, tracer)
	if err != nil {
		panic(err)
	}
	return db
}

func main() {
	/* init sentry */
	sentryErr := sentry.Init(sentry.ClientOptions{
		Dsn: SENTRY_DSN,
	})
	/* init tracing*/
	tracer, closer := _util_tracing.Init("go-clean-template")
	defer closer.Close()
	opentracing.SetGlobalTracer(tracer)

	psqlClient := sqlDBWithTracing(PSQL_DATABASE_URL, tracer)
	defer psqlClient.GetClient().Close()

	/* init grpc */
	server := grpc.NewServer(
		grpc.UnaryInterceptor(
			otgrpc.OpenTracingServerInterceptor(tracer),
		),
		grpc.StreamInterceptor(
			otgrpc.OpenTracingStreamServerInterceptor(tracer),
		),
	)
	defer server.GracefulStop()

	e := echo.New()
	e.HTTPErrorHandler = helperMiddl.SentryCapture(e)
	helperRoute.RegisterVersion(e)

	e.Use(echoMiddL.Logger())
	e.Use(echoMiddL.Recover())
	e.Use(sentryecho.New(sentryecho.Options{Repanic: true}))

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	middL := myMiddL.InitMiddleware(JWT_SECRET)
	e.Use(echoMiddL.Recover())
	e.Use(echoMiddL.CORSWithConfig(echoMiddL.CORSConfig{
		Skipper:      echoMiddL.DefaultSkipper,
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete},
	}))
	e.Use(middL.InitContextIfNotExists)
	e.Use(middL.InputForm)
	e.Use(middL.SetTracer)

	r := route.NewRoute(e, middL)

	/* repository */
	orgRepo := _organize_repository.NewPsqlOrganizeRepository(psqlClient)

	/* usecase */
	orgUs := _organize_usecase.NewOrganizeUsecase(orgRepo)

	/* handler */
	orgHandler := _organize_handler.NewOrganizeHandler(orgUs)

	/* gprc handler */
	orgGRPCHandler := _organize_grpc.NewGRPCOrganizeHandler(orgUs)

	/* validate */
	orgValidation := _organize_validator.Validation{}

	/* inject route */
	r.RegisterOrganization(orgHandler, orgValidation)

	/* inject grpc route */
	grpcRoute := route.NewGRPCRoute(server)
	grpcRoute.RegisterOrganize(orgGRPCHandler)

	/* serve gprc */
	go func() {
		if r := recover(); r != nil {
			fmt.Println("error on start grpc server: ", r.(error))
		}
		startGRPCServer(server)
	}()

	/* serve echo */
	port := fmt.Sprintf(":%s", APP_PORT)
	if sentryErr == nil {
		sentry.CaptureException(e.Start(port))
	} else {
		e.Logger.Fatal(e.Start(port))
	}
}

func startGRPCServer(server *grpc.Server) {
	listen, err := net.Listen("tcp", fmt.Sprintf(":%s", GRPC_PORT))
	if err != nil {
		panic("failed to listen: " + err.Error())
	}

	/* serve grpc */
	fmt.Println(fmt.Sprintf("Start grpc Server [::%s]", GRPC_PORT))
	if err := server.Serve(listen); err != nil {
		panic(err)
	}
}
