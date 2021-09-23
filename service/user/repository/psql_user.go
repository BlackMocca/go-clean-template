package repository

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"git.innovasive.co.th/backend/psql"
	"github.com/BlackMocca/go-clean-template/constants"
	myHelper "github.com/BlackMocca/go-clean-template/helper"
	"github.com/BlackMocca/go-clean-template/models"
	"github.com/BlackMocca/go-clean-template/orm"
	"github.com/BlackMocca/go-clean-template/service/user"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/cast"
)

type psqlUserRepository struct {
	db *psql.Client
}

func NewPsqlUserRepository(dbcon *psql.Client) user.UserRepository {
	return &psqlUserRepository{
		db: dbcon,
	}
}

func (p psqlUserRepository) whereCond(args *sync.Map) []string {
	var conds = []string{}

	if v, ok := args.Load("user_type_id"); ok && v != nil {
		sql := fmt.Sprintf("user_types.id::text = '%s'", cast.ToString(v))
		conds = append(conds, sql)
	}

	return conds
}

func (p psqlUserRepository) FetchAll(args *sync.Map) ([]*models.User, error) {
	var conds = p.whereCond(args)
	var where string
	if len(conds) > 0 {
		where = "WHERE " + strings.Join(conds, " AND ")
	}
	sql := fmt.Sprintf(`
		SELECT 
			%s,
			%s
		FROM users
		JOIN
			user_types
		ON
			users.user_type_id = user_types.id
		%s
	`,
		orm.GetSelector(models.User{}),
		orm.GetSelector(models.UserType{}),
		where,
	)

	myHelper.Println(sql)

	rows, err := p.db.GetClient().Queryx(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	joinField := []string{models.FIELD_FK_USER_TYPE}
	users, err := p.orm(rows, joinField)
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
		JOIN
			user_types
		ON
			users.user_type_id = user_types.id
		WHERE users.id=%d
	`,
		orm.GetSelector(models.User{}),
		id,
	)

	myHelper.Println(sql)

	rows, err := p.db.GetClient().Queryx(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	joinField := []string{models.FIELD_FK_USER_TYPE}
	users, err := p.orm(rows, joinField)
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
	var mapper, err = orm.NewRowsScan(rows)
	if err != nil {
		return nil, err
	}

	if mapper.TotalRows() > 0 {
		for _, row := range mapper.RowsValues() {
			var user = new(models.User)
			user, err := orm.OrmUser(user, mapper, row, joinField)
			if err != nil {
				return nil, err
			}
			if user != nil {
				exists, err := orm.IsDuplicateByPK(users, user)
				if err != nil {
					return nil, err
				}
				if !exists {
					users = append(users, user)
				}
			}
		}
	}

	if len(users) > 0 {
		for index, _ := range users {
			if err := orm.OrmUserRelation(users[index], mapper, joinField); err != nil {
				return nil, err
			}
		}
	}

	return users, nil
}
