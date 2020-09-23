package repository_test

import (
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v2"

	"github.com/BlackMocca/go-clean-template/models"
	_user_repository "github.com/BlackMocca/go-clean-template/service/user/repository"
	"github.com/jmoiron/sqlx"
)

func TestCreateSuccess(t *testing.T) {
	now := time.Now()
	ar := &models.User{
		Email:     "abc@gmail.com",
		Firstname: "Kongitat",
		Lastname:  "Monkol",
		Age:       23,
		CreatedAt: &now,
		UpdatedAt: &now,
	}

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	defer db.Close()

	mock.ExpectBegin()

	query := `^INSERT (.+)`
	prep := mock.ExpectPrepare(query)
	prep.ExpectExec().WithArgs(ar.Email, ar.Firstname, ar.Lastname, ar.Age, ar.CreatedAt, ar.UpdatedAt).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	a := _user_repository.NewPsqlUserRepository(sqlxDB)

	err = a.Create(ar)
	if err != nil {
		log.Panic(err)
	}
	assert.NoError(t, err)
}
