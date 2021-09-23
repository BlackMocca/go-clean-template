package repository

import (
	"errors"
	"fmt"

	"git.innovasive.co.th/backend/psql"
	"github.com/BlackMocca/go-clean-template/constants"
	myHelper "github.com/BlackMocca/go-clean-template/helper"
	"github.com/BlackMocca/go-clean-template/models"
	"github.com/BlackMocca/go-clean-template/orm"
	"github.com/BlackMocca/go-clean-template/service/user"
	"github.com/jmoiron/sqlx"
)

type psqlUserRepository struct {
	db *psql.Client
}

func NewPsqlUserRepository(dbcon *psql.Client) user.UserRepository {
	return &psqlUserRepository{
		db: dbcon,
	}
}

func (p psqlUserRepository) FetchAll() ([]*models.User, error) {
	sql := fmt.Sprintf(`
		SELECT 
			%s
		FROM users
	`,
		models.UserSelector,
	)

	myHelper.Println(sql)

	rows, err := p.db.GetClient().Queryx(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users, err := p.orm(rows, nil)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (p psqlUserRepository) FetchOneById(id int64) (*models.User, error) {
	sql := fmt.Sprintf(`
		SELECT 
			%s
		FROM users
		WHERE users.id=%d
	`,
		models.UserSelector,
		id,
	)

	myHelper.Println(sql)

	rows, err := p.db.GetClient().Queryx(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users, err := p.orm(rows, nil)
	if err != nil {
		return nil, err
	}
	if len(users) > 0 {
		return users[0], nil
	}

	return nil, nil
}

func (p psqlUserRepository) Create(user *models.User) error {
	tx, err := p.db.GetClient().Begin()
	if err != nil {
		return err
	}

	sql := `
		INSERT INTO users(id,email,firstname,lastname,age, user_type_id ,created_at,updated_at,deleted_at)
		VALUES ($1::uuid, $2::text, $3::text, $4::text, $5::numeric, $6::uuid, $7::timestamp, $8::timestamp, NULL)
	`

	myHelper.Println(sql)

	stmt, err := tx.Prepare(sql)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		user.Id,
		user.Email,
		user.Firstname,
		user.Lastname,
		user.Age,
		user.UserTypeId,
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		if err.Error() == constants.ERROR_DUPLICATE_EMAIL {
			return errors.New(constants.ERROR_DUPLICATE_EMAIL_MESSAGE)
		}
		return err
	}

	return tx.Commit()
}

func (p psqlUserRepository) orm(rows *sqlx.Rows, joinField []string) ([]*models.User, error) {
	var users = make([]*models.User, 0)

	for rows.Next() {
		var user = new(models.User)
		user, err := orm.OrmUser(user, rows, joinField)
		if err != nil {
			return nil, err
		}
		if user != nil {
			users = append(users, user)
		}
	}

	return users, nil
}
