package orm_test

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v2"

	helperModel "git.innovasive.co.th/backend/models"
	"github.com/BlackMocca/go-clean-template/models"
	"github.com/BlackMocca/go-clean-template/orm"
	"github.com/jmoiron/sqlx"
)

func TestOrmUser(t *testing.T) {
	now := helperModel.NewTimestampFromTime(time.Now())
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
	}

	rows := sqlmock.NewRows([]string{"id", "email", "firstname", "lastname", "age", "created_at", "updated_at", "deleted_at"})
	for _, item := range mockUsers {
		rows.AddRow(item.ID, item.Email, item.Firstname, item.Lastname, item.Age,
			item.CreatedAt, item.UpdatedAt, item.DeletedAt)
	}

	query := fmt.Sprintf("SELECT %s FROM users", models.UserSelector)
	mock.ExpectQuery(query).WillReturnRows(rows)

	sqlxRows, err := sqlxDB.Queryx(query)
	log.Println(sqlxRows)
	defer sqlxRows.Close()

	for sqlxRows.Next() {
		user := new(models.User)
		user, err = orm.OrmUser(user, sqlxRows, nil)
		assert.NoError(t, err)
		if assert.NotEmpty(t, user) {
			assert.Equal(t, user.ID, mockUsers[0].ID)
			assert.Equal(t, user.Email, mockUsers[0].Email)
			assert.Equal(t, user.Firstname, mockUsers[0].Firstname)
			assert.Equal(t, user.Lastname, mockUsers[0].Lastname)
			assert.Equal(t, user.Age, mockUsers[0].Age)
			assert.Equal(t, user.CreatedAt, mockUsers[0].CreatedAt)
			assert.Equal(t, user.UpdatedAt, mockUsers[0].UpdatedAt)
		}
	}
}
