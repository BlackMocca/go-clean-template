package config

import (
	"log"
	"os"

	"github.com/go-pg/pg/v9"
	_ "github.com/lib/pq"
)

func NewPsqlConnection() *pg.DB {
	PsqlConnectionStr, has := os.LookupEnv("PSQL_DATABASE_URL")
	if !has {
		log.Fatal("PSQL_DATABASE_URL on env not found")
	}
	option, err := pg.ParseURL(PsqlConnectionStr)
	if err != nil {
		log.Fatal(err)
	}
	db := pg.Connect(option)

	return db
}
