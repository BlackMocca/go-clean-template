// main.go — Echo server with API Key authentication middleware (Manual Route Registration).
//
// This file demonstrates how to wire together:
//   - oapi-codegen generated handlers and route wrappers
//   - Custom API Key authentication middleware (header or query param)
//   - Manual route registration separating public vs protected endpoints
//
// Usage:
//
//	go run .
//
// Authentication:
//   - Header:    X-API-Key: your-key-here
//   - Query:     ?apikey=your-key-here
//
// Test keys (development):
//   - exampleapikey
package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"

	"example/middleware"
	"example/migrations"
	"example/route"
	"example/service/user"
	"example/service/web"
	api "example/service/web" // import generated code จาก service/web/ package

	"github.com/BlackMocca/utils/connectors/psql"
	"github.com/labstack/echo/v4"
	"github.com/spf13/cast"

	_ "github.com/lib/pq" // PostgreSQL driver

	_ "github.com/golang-migrate/migrate/v4/source/file"
	echoMiddL "github.com/labstack/echo/v4/middleware"
)

type Config struct {
	APIKey      string // allowed API key value for request validation
	Port        int    // port the HTTP server listens on
	DatabaseURL string // postgres database url
}

func lookupEnv(key string, defaultVal string) string {
	val := os.Getenv(key)
	if val != "" {
		return val
	}
	return defaultVal
}

// loadConfig reads configuration from environment variables.
// Falls back to sensible defaults when ENV vars are not set (useful for local development).
func loadConfig() (*Config, error) {
	config := &Config{
		APIKey:      lookupEnv("APIKEY", ""),
		Port:        cast.ToInt(lookupEnv("PORT", "3001")),
		DatabaseURL: lookupEnv("DATABASE_URL", ""),
	}

	return config, nil
}

// ──────────── Database Connect ────────────
func connectDatabase(ctx context.Context, config *Config) *psql.Client {
	client, err := psql.NewConnectionWithTracing(ctx, config.DatabaseURL, psql.Postgres, nil)
	if err != nil {
		log.Fatalf("failed to connect database: %s", err.Error())
	}
	return client
}

// ──────────── Migrator Runner ────────────

func migrateRunner(db *psql.Client) {
	var STOP_ON_MIGRATION_FAILED = cast.ToBool(lookupEnv("STOP_ON_MIGRATION_FAILED", "true"))

	relPath := filepath.Join(".", "migrations", "database", "postgres")
	absPath, err := filepath.Abs(relPath)
	if err != nil {
		panic(err)
	}
	config, err := migrations.NewConfig(db.GetConnectionURI(), "public", absPath)
	if err != nil {
		err = fmt.Errorf("failed to initial migration: %w", err)
		if STOP_ON_MIGRATION_FAILED {
			panic(err)
		} else {
			slog.Error(err.Error())
		}
	}

	if err := config.Migrate(); err != nil {
		err = fmt.Errorf("failed to migration: %w", err)
		if STOP_ON_MIGRATION_FAILED {
			panic(err)
		} else {
			slog.Error(err.Error())
		}
	}
}

func main() {
	ctx := context.Background()
	// 1. Load configuration (reads API keys from ENV or uses auto-generated test keys)
	config, err := loadConfig()
	if err != nil {
		log.Fatalf("❌ Failed to load config: %v", err)
	}

	fmt.Println("🚀 Starting Example API Server")
	fmt.Printf("🔑 Registered API keys: %d keys loaded\n", len(config.APIKey))
	fmt.Println()

	// connect database
	psqlClient := connectDatabase(ctx, config)
	// Migrate database
	migrateRunner(psqlClient)

	// 2. Create Echo HTTP server instance
	middl := middleware.NewClient(middleware.APIKeyAuthDefault("X-API-KEY", config.APIKey))
	e := echo.New()
	e.HideBanner = false
	e.Use(middl.InitContextIfNotExists)
	e.Use(echoMiddL.CORS())
	e.Use(echoMiddL.RequestID())
	e.Use(echoMiddL.AddTrailingSlash())
	e.Use(echoMiddL.Recover())
	e.Use(middl.TransformResponse)
	e.Use(middl.CreateOpenapiMiddleware(web.GetSwagger))

	// 3. init service
	// repository
	userRepo := user.NewPsqlRepository(psqlClient)

	// usecase
	userUs := user.NewUserUsecase(userRepo)

	// 4. Register routes manually — public and protected endpoints are clearly separated
	handler := api.NewHandler(userUs)
	wrapper := api.ServerInterfaceWrapper{Handler: handler}

	r := route.NewRoute(e, middl)

	r.RegisterPublicRoutes(wrapper)
	r.RegisterProtectedRoutes(wrapper)

	// 5. Start HTTP server (blocking)
	addr := fmt.Sprintf(":%d", config.Port)
	log.Printf("📡 Server listening on http://0.0.0.0%s\n", addr)
	e.Logger.Fatal(e.Start(addr))
}
