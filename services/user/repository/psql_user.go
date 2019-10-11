package repository

import (
	driver "database/sql/driver"
)

type psqlUserRepository struct {
	conn *driver.Conn
}

func newPsqlUserRepository(connstr string)
