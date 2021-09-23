package main

import (
	"git.innovasive.co.th/backend/psql"

	"github.com/BlackMocca/go-clean-template/config"
	_ "github.com/BlackMocca/go-clean-template/integration"
	"github.com/BlackMocca/go-clean-template/server"
)

func sqlDB(con string) *psql.Client {
	db, err := psql.NewPsqlConnection(con)
	if err != nil {
		panic(err)
	}
	return db
}

func GetMainServer() *server.Server {
	psqlDB := sqlDB(config.PSQL_DATABASE_URL)

	return &server.Server{
		APP_LOGGER:   config.APP_LOGGER,
		APP_PORT:     config.APP_PORT,
		JWT_SECRET:   config.JWT_SECRET,
		GRPC_PORT:    config.GRPC_PORT,
		GRPC_TIMEOUT: config.GRPC_TIMEOUT,
		SENTRY_DSN:   config.SENTRY_DSN,
		PsqlDB:       psqlDB,
	}
}

func main() {
	serv := GetMainServer()
	defer serv.PsqlDB.GetClient().Close()
	serv.Start()
}
