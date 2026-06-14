package user_test

import (
	"context"
	"example/migrations"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/BlackMocca/utils/connectors/psql"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type PostgresRepositorySuite struct {
	suite.Suite

	Ctx context.Context

	PostgresqlContainer testcontainers.Container
	client              *psql.Client
	dbName              string
	dbUser              string
	dbPassword          string
}

func (s *PostgresRepositorySuite) SetupSuite() {
	s.Ctx = context.Background()
	// Start Container
	if err := s.StartContainer(s.Ctx); err != nil {
		s.Require().NoError(err)
	}
	// Connect Database
	if err := s.ConnectDatabase(s.Ctx); err != nil {
		s.Require().NoError(err)
	}
	// Run Migration
	if err := s.RunMigrations(s.client.GetConnectionURI()); err != nil {
		s.Require().NoError(err)
	}

}

func (s *PostgresRepositorySuite) TearDownSuite() {
	// Stop Container
	s.StopContainer()
}

func (s *PostgresRepositorySuite) SetupTest() {
	// Truncate Tables
	s.TruncateTables(s.Ctx)
}

func (s *PostgresRepositorySuite) TearDownTest() {
	// Optional
}

func (s *PostgresRepositorySuite) BeforeTest(suiteName string, testName string) {
	// Optional
}

func (s *PostgresRepositorySuite) AfterTest(suiteName string, testName string) {
	// Optional
}

func (s *PostgresRepositorySuite) StartContainer(ctx context.Context) error {
	var dbName = "example"
	var dbUser = "user"
	var dbPassword = "password"

	req := testcontainers.ContainerRequest{
		Image: "postgres:16-alpine",

		Env: map[string]string{
			"POSTGRES_DB":       dbName,
			"POSTGRES_USER":     dbUser,
			"POSTGRES_PASSWORD": dbPassword,
		},

		ExposedPorts: []string{"5432/tcp"},

		WaitingFor: wait.ForListeningPort("5432/tcp"),

		Tmpfs: map[string]string{
			"/var/lib/postgresql/data": "rw",
		},
	}

	container, err := testcontainers.GenericContainer(
		ctx,
		testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		},
	)
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return err
	}

	// container log
	logs, err := container.Logs(ctx)
	if err != nil {
		return err
	}
	defer logs.Close()

	data, err := io.ReadAll(logs)
	if err != nil {
		return err
	}
	fmt.Println("---------------------------Postgres Container Logs-------------------------------")
	fmt.Println(string(data))
	fmt.Println("---------------------------Postgres Container Logs-------------------------------")

	s.PostgresqlContainer = container
	s.dbUser = dbUser
	s.dbPassword = dbPassword
	s.dbName = dbName
	return nil
}

func (s *PostgresRepositorySuite) StopContainer() error {
	defer func() {
		if err := testcontainers.TerminateContainer(s.PostgresqlContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	return nil
}

func (s *PostgresRepositorySuite) ConnectDatabase(ctx context.Context) error {
	host, err := s.PostgresqlContainer.ContainerIP(ctx)
	if err != nil {
		return err
	}
	connStr := fmt.Sprintf("postgres://%s:%s@%s:5432/%s?sslmode=disable",
		s.dbUser,
		s.dbPassword,
		host,
		s.dbName,
	)
	client, err := psql.NewConnection(ctx, connStr)
	if err != nil {
		return err
	}

	s.client = client
	return nil
}

func (s *PostgresRepositorySuite) RunMigrations(databaseDSN string) error {
	// path relative start with at file test
	config, err := migrations.NewConfig(s.client.GetConnectionURI(), "public", "../../migrations/database/postgres")
	if err != nil {
		return err
	}
	return config.Migrate()
}

func (s *PostgresRepositorySuite) TruncateTables(ctx context.Context) error {
	var tables []string

	err := s.client.GetClient().SelectContext(ctx, &tables, `
		SELECT tablename
		FROM pg_tables
		WHERE schemaname = 'public'
	`)
	if err != nil && err != pgx.ErrNoRows {
		return err
	}

	if len(tables) == 0 {
		return nil
	}

	query := fmt.Sprintf(
		"TRUNCATE TABLE %s RESTART IDENTITY CASCADE",
		strings.Join(tables, ", "),
	)

	_, exerr := s.client.GetClient().ExecContext(ctx, query)
	return exerr
}
