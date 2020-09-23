package config

import (
	"github.com/jmoiron/sqlx"
	pg "github.com/lib/pq"
)

func NewPsqlConnection(connectionStr string) (*sqlx.DB, error) {
	addr, err := pg.ParseURL(connectionStr)
	if err != nil {
		return nil, err
	}

	db, err := sqlx.Connect("postgres", addr)
	if err != nil {
		return nil, err
	}

	return db, nil
}
