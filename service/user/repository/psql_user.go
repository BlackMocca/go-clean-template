package repository

import (
	"fmt"
	"log"

	"github.com/BlackMocca/go-clean-template/models"
	"github.com/BlackMocca/go-clean-template/orm"
	"github.com/BlackMocca/go-clean-template/service/user"
	"github.com/jmoiron/sqlx"
)

type psqlUserRepository struct {
	db *sqlx.DB
}

func NewPsqlUserRepository(dbcon *sqlx.DB) user.PsqlUserRepositoryInf {
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

	log.Println(sql)

	rows, err := p.db.Queryx(sql)
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

	log.Println(sql)

	rows, err := p.db.Queryx(sql)
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
	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	sql := `
		INSERT INTO users(id,email,firstname,lastname,age,created_at,updated_at,deleted_at)
		VALUES (nextval('users_id_seq'), $1::text, $2::text, $3::text, $4::numeric, $5::timestamp, $6::timestamp, NULL)
	`

	stmt, err := tx.Prepare(sql)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		user.Email,
		user.Firstname,
		user.Lastname,
		user.Age,
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
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
