package config

import (
	"database/sql"

	"github.com/spf13/viper"
)

func (c *Config) NewConnectPsql(*viper.Viper) *sql.DB {
	connStr := "postgres://postgres:postgres@localhost/pqgotest?sslmode=false"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	return db
}
