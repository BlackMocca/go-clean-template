// build integration
package integration_test

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"git.innovasive.co.th/backend/psql"
	"github.com/BlackMocca/go-clean-template/config"
	"github.com/BlackMocca/go-clean-template/server"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

/* @doc
https://github.com/stretchr/testify/blob/master/suite/interfaces.go
*/

var (
	containerMigrateDir    = "/migrations"
	postgresMigrateDir     = "/migrations/database/postgres"
	seedMasterMigrateDir   = "/migrations/database/postgres/seed/master"
	seedDataTestMigrateDir = func(story string) string {
		return fmt.Sprintf("/migrations/database/postgres/seed/%s", story)
	}
)

var (
	postgresExposePort     = "5432"
	postgresExposeProtocal = "tcp"
	postgresUser           = "postgres"
	postgresPassword       = "postgres"
	postgressDatabase      = "app_test"
	postgresMigrateURI     = func(user, pass, host, port, database string) string {
		return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, pass, host, port, database)
	}
)

var (
	getStoryName = func(name string) string {
		reg := regexp.MustCompile(`MIGRATE_STORY_\d+`)
		if reg.MatchString(name) {
			return strings.ReplaceAll(strings.ReplaceAll(strings.ToLower(reg.FindString(name)), "_", "-"), "migrate-", "")
		}
		return ""
	}
)

type E2eTestSuite struct {
	suite.Suite
	psqlClient      *psql.Client
	psqlDbContainer tc.Container
}

func TestE2ETestSuite(t *testing.T) {
	suite.Run(t, &E2eTestSuite{})
}

func (e *E2eTestSuite) SetupSuite() {
	/* init docker postgres for setup */
	ctx := context.Background()

	req := tc.ContainerRequest{
		FromDockerfile: tc.FromDockerfile{
			Context:    "../docker-script/postgres",
			Dockerfile: "./Dockerfile",
		},
		ExposedPorts: []string{fmt.Sprintf("%s/%s", postgresExposePort, postgresExposeProtocal)},
		Env: map[string]string{
			"POSTGRES_USER":     postgresUser,
			"POSTGRES_PASSWORD": postgresPassword,
			"POSTGRES_DB":       postgressDatabase,
		},
		AutoRemove: true,
		Name:       "psql_db_test",
		WaitingFor: wait.ForLog("PostgreSQL init process complete; ready for start up"),
		BindMounts: map[string]string{
			config.GetPath("migrations"): containerMigrateDir,
		},
	}

	postgresC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		panic(err)
	}

	host, err := postgresC.Host(ctx)
	if err != nil {
		panic(err)
	}

	port, err := postgresC.MappedPort(ctx, nat.Port(postgresExposePort))
	if err != nil {
		panic(err)
	}

	/* connect db */
	psqlURL := postgresMigrateURI(postgresUser, postgresPassword, host, port.Port(), postgressDatabase)
	psqlDB, err := psql.NewPsqlConnection(psqlURL)
	if err != nil {
		panic(fmt.Errorf("fail to connection: %s", err.Error()))
	}

	serv := &server.Server{
		APP_LOGGER:   config.APP_LOGGER,
		APP_PORT:     config.APP_PORT,
		JWT_SECRET:   config.JWT_SECRET,
		GRPC_PORT:    config.GRPC_PORT,
		GRPC_TIMEOUT: config.GRPC_TIMEOUT,
		SENTRY_DSN:   config.SENTRY_DSN,
		PsqlDB:       psqlDB,
	}

	go serv.Start()
	time.Sleep(time.Duration(2 * time.Second))

	e.psqlDbContainer = postgresC
	e.psqlClient = psqlDB
}

func (e *E2eTestSuite) SetupTest() {
	host, err := e.psqlDbContainer.Host(context.Background())
	if err != nil {
		e.T().Error(err)
	}

	/* migrate table */
	migrateDir := postgresMigrateDir
	migrate := fmt.Sprintf(`migrate -database %s -path %s up`, postgresMigrateURI(postgresUser, postgresPassword, host, postgresExposePort, postgressDatabase), migrateDir)
	cmd := strings.Split(migrate, " ")
	status, err := e.psqlDbContainer.Exec(context.Background(), cmd)
	if err != nil {
		e.T().Error(err)
	}

	switch status {
	case 125:
		e.T().Error("problem with Docker itself")
	case 127:
		e.T().Error("contained command could not be invoked")
	}

	fmt.Println("migrate table success!!!!")

	/* migrate seed master */
	migrateDir = seedMasterMigrateDir
	migrate = fmt.Sprintf(`migrate -database %s -path %s seed-up`, postgresMigrateURI(postgresUser, postgresPassword, host, postgresExposePort, postgressDatabase), migrateDir)
	cmd = strings.Split(migrate, " ")
	fmt.Println(migrate)
	status, err = e.psqlDbContainer.Exec(context.Background(), cmd)
	if err != nil {
		e.T().Error(err)
	}

	switch status {
	case 125:
		e.T().Error("problem with Docker itself")
	case 127:
		e.T().Error("contained command could not be invoked")
	}

	fmt.Println("migrate seed master success!!!!")
}

func (e *E2eTestSuite) TearDownSuite() {
	/* down docker compose */
	if err := e.psqlDbContainer.Terminate(context.Background()); err != nil {
		e.T().Error(err)
	}
}

func (e *E2eTestSuite) TearDownTest() {
	/* down data from database */
	host, err := e.psqlDbContainer.Host(context.Background())
	if err != nil {
		e.T().Error(err)
	}

	migrate := fmt.Sprintf(`migrate -database %s -path %s down`, postgresMigrateURI(postgresUser, postgresPassword, host, postgresExposePort, postgressDatabase), postgresMigrateDir)
	cmd := strings.Split(migrate, " ")
	status, err := e.psqlDbContainer.Exec(context.Background(), cmd)
	if err != nil {
		e.T().Error(err)
	}

	switch status {
	case 125:
		e.T().Error("problem with Docker itself")
	case 127:
		e.T().Error("contained command could not be invoked")
	}

	fmt.Println("delete table success!!!!")
}

func (e *E2eTestSuite) BeforeTest(suiteName, testName string) {
	/* seed data from test story */
	fmt.Println()
	fmt.Println("Starting Test Name: ", testName)
	fmt.Println()
	host, err := e.psqlDbContainer.Host(context.Background())
	if err != nil {
		e.T().Error(err)
	}

	testName = getStoryName(testName)
	if testName != "" {
		fmt.Println("Start Seed Up on story", testName)
		migrateDir := seedDataTestMigrateDir(testName)
		migrate := fmt.Sprintf(`migrate -database %s -path %s seed-up`, postgresMigrateURI(postgresUser, postgresPassword, host, postgresExposePort, postgressDatabase), migrateDir)
		cmd := strings.Split(migrate, " ")
		status, err := e.psqlDbContainer.Exec(context.Background(), cmd)
		if err != nil {
			e.T().Error(err)
		}

		switch status {
		case 125:
			e.T().Error("problem with Docker itself")
		case 127:
			e.T().Error("contained command could not be invoked")
		}
	}

}

func (e *E2eTestSuite) AfterTest(suiteName, testName string) {
	/* delete data from table */
	host, err := e.psqlDbContainer.Host(context.Background())
	if err != nil {
		e.T().Error(err)
	}

	testName = getStoryName(testName)
	if testName != "" {
		fmt.Println("Start Seed Down on story", testName)
		migrateDir := seedDataTestMigrateDir(testName)
		migrate := fmt.Sprintf(`migrate -database %s -path %s seed-down`, postgresMigrateURI(postgresUser, postgresPassword, host, postgresExposePort, postgressDatabase), migrateDir)
		cmd := strings.Split(migrate, " ")
		status, err := e.psqlDbContainer.Exec(context.Background(), cmd)
		if err != nil {
			e.T().Error(err)
		}

		switch status {
		case 125:
			e.T().Error("problem with Docker itself")
		case 127:
			e.T().Error("contained command could not be invoked")
		}
	}
}
