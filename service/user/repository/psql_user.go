package repository

import (
	"github.com/go-pg/pg/v9"
	"github.com/BlackMocca/go-clean-template/service/user"
)

type psqlUserRepository struct {
	db *pg.DB
}

func NewPsqlUserRepository(dbcon *pg.DB) user.PsqlUserRepositoryInf {
	return &psqlUserRepository{
		db: dbcon,
	}
}
