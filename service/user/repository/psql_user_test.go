package repository

import (
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v2"

	helperModel "git.innovasive.co.th/backend/models"
	"github.com/BlackMocca/go-clean-template/models"
	"github.com/jmoiron/sqlx"
)

var (
	now = helperModel.NewTimestampFromTime(time.Now())
)

func TestFetchAll(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	defer db.Close()

	mockUsers := []*models.User{
		&models.User{
			ID: 1, Email: "abc@gmail.com", Firstname: "Kongitat", Lastname: "Monkol", Age: 30,
			CreatedAt: &now, UpdatedAt: &now, DeletedAt: nil,
		},
		&models.User{
			ID: 2, Email: "xyz@gmail.com", Firstname: "Kittitat", Lastname: "Monkolchart", Age: 25,
			CreatedAt: &now, UpdatedAt: &now, DeletedAt: nil,
		},
		&models.User{
			ID: 3, Email: "qwerty@gmail.com", Firstname: "Teeradet", Lastname: "Phondetparinya", Age: 25,
			CreatedAt: &now, UpdatedAt: &now, DeletedAt: &now,
		},
	}

	rows := sqlmock.NewRows([]string{"id", "email", "firstname", "lastname", "age", "created_at", "updated_at", "deleted_at"})
	for _, item := range mockUsers {
		rows.AddRow(item.ID, item.Email, item.Firstname, item.Lastname, item.Age,
			item.CreatedAt, item.UpdatedAt, item.DeletedAt)
	}

	query := "^SELECT (.+) FROM users"
	mock.ExpectQuery(query).WillReturnRows(rows)

	a := NewPsqlUserRepository(sqlxDB)
	list, err := a.FetchAll()
	assert.NoError(t, err)
	assert.Len(t, list, 2)
}

func TestFetchOneById(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	defer db.Close()

	mockUser := &models.User{
		ID: 1, Email: "qwerty@gmail.com", Firstname: "Teeradet", Lastname: "Phondetparinya", Age: 25,
		CreatedAt: &now, UpdatedAt: &now, DeletedAt: nil,
	}
	a := NewPsqlUserRepository(sqlxDB)
	rows := sqlmock.NewRows([]string{"id", "email", "firstname", "lastname", "age", "created_at", "updated_at", "deleted_at"})
	rows.AddRow(mockUser.ID, mockUser.Email, mockUser.Firstname, mockUser.Lastname, mockUser.Age,
		mockUser.CreatedAt, mockUser.UpdatedAt, mockUser.DeletedAt)

	query := "^SELECT (.+) FROM users"
	mock.ExpectQuery(query).WillReturnRows(rows)

	user, err := a.FetchOneById(int64(3))
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, user, mockUser)
}

func TestCreate(t *testing.T) {
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

	a := NewPsqlUserRepository(sqlxDB)

	err = a.Create(ar)
	if err != nil {
		log.Panic(err)
	}
	assert.NoError(t, err)
}
