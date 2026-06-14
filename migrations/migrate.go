package migrations

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/BlackMocca/utils/connectors/psql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type Config struct {
	schema string
	path   string

	client *psql.Client
}

func getMigrationPath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}

	return filepath.Clean(path)
}

func NewConfig(database string, schema string, path string) (Config, error) {
	client, err := psql.NewConnection(context.Background(), database)
	if err != nil {
		return Config{}, fmt.Errorf("failed: %s", err.Error())
	}

	return Config{
		schema: schema,
		path:   getMigrationPath(path),
		client: client,
	}, nil
}

func (config Config) Migrate() error {
	defer config.client.Close()

	fmt.Println(`[INFO] create schema ...`)
	if err := config.createSchemaIfNotExists(config.schema); err != nil {
		return err
	}

	fmt.Println(`[INFO] starting migration...`)
	if err := config.migrateDatabase(config.schema, "schema_migrations"); err != nil {
		return err
	}

	fmt.Println(`[INFO] starting seed data...`)
	config.path = filepath.Join(config.path, "master")
	if err := config.migrateDatabase(config.schema, "seed_migrations"); err != nil && !strings.Contains(err.Error(), "no such file or directory") {
		return err
	}

	fmt.Println(`[INFO] migration successful!!!!!`)
	return nil
}

func (config Config) createSchemaIfNotExists(schema string) error {
	_, err := config.client.GetClient().Exec(fmt.Sprintf(`CREATE SCHEMA IF NOT EXISTS "%s"`, schema))
	fmt.Printf("[INFO] Create Schema \"%s\" successful \n", schema)
	return err
}

func (config *Config) migrateDatabase(schema string, migrationTable string) error {
	_, err := config.client.GetClient().DB.Exec(fmt.Sprintf("SET search_path TO \"%s\"", schema))
	if err != nil {
		return fmt.Errorf("failed to set search_path: %w", err)
	}

	driver, err := postgres.WithInstance(config.client.GetClient().DB, &postgres.Config{
		SchemaName:      schema,
		MigrationsTable: migrationTable,
	})
	if err != nil {
		return fmt.Errorf("failed to create database driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+filepath.Join(config.path),
		"",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration failed: %w", err)
	}

	version, _, err := m.Version()
	if err != nil {
		return err
	}

	fmt.Printf("[INFO] migration version %d successful \n", version)

	return nil
}
