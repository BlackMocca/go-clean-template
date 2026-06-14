// Package user provides a repository interface and in-memory implementation
// for the example unit testing demonstration.
package user

import (
	"context"
	"example/models"
	"fmt"
	"strings"

	"github.com/BlackMocca/utils/connectors/psql"
	"github.com/google/uuid"
	"github.com/guregu/null/v6/zero"
	"github.com/jmoiron/sqlx"
)

type psqlRepository struct {
	client *psql.Client
}

func NewPsqlRepository(client *psql.Client) Repository {
	return &psqlRepository{client: client}
}

// FindByID retrieves a user by their unique ID. Returns an error if not found.
func (p psqlRepository) FindByID(ctx context.Context, idOrEmail string) (*models.User, error) {
	uid, _ := uuid.Parse(idOrEmail)
	sql := `
		SELECT 
			users.*,
			COALESCE(
				json_agg(row_to_json(permissions)) FILTER (WHERE permissions.id IS NOT NULL),
				'[]'::json
			) AS permissions
		FROM 
			users
		LEFT JOIN
			user_permissions
		ON
			users.id = user_permissions.user_id
		LEFT JOIN
			permissions
		ON
			permissions.id = user_permissions.permission_id
		WHERE 
			users.deleted_at IS NULL
			AND
			(users.id = ? or users.email = ?)
		GROUP BY users.id
	ORDER BY users.id ASC
	`
	sql = sqlx.Rebind(sqlx.DOLLAR, sql)
	stmt, err := p.client.GetClient().PreparexContext(ctx, sql)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Unsafe().QueryxContext(ctx, uid, idOrEmail)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users, err := p.orm(ctx, rows, nil)
	if err != nil {
		return nil, err
	}

	if len(users) > 0 {
		return users[0], nil
	}

	return nil, nil
}

func (p psqlRepository) upsertUser(ctx context.Context, tx *sqlx.Tx, user *models.User) error {
	sql := `
		INSERT INTO users (id, name, email, role, created_at, updated_at, deleted_at)
		VALUES (?, ?, ?, ?::user_role, ?,?, NULL)
		ON CONFLICT (id)
		DO UPDATE SET
			email = EXCLUDED.email,
			name = EXCLUDED.name,
			updated_at = EXCLUDED.updated_at;
	`

	sql = sqlx.Rebind(sqlx.DOLLAR, sql)
	stmt, err := tx.PreparexContext(ctx, sql)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, merr := stmt.ExecContext(ctx,
		user.ID,
		user.Name,
		user.Email,
		user.Role,
		user.CreatedAt,
		user.UpdatedAt,
	)
	return merr
}

func (p psqlRepository) deleteUserPermission(ctx context.Context, tx *sqlx.Tx, userId uuid.UUID) error {
	sql := `
		DELETE FROM user_permissions
		WHERE
			user_id = ?
	`

	sql = sqlx.Rebind(sqlx.DOLLAR, sql)
	stmt, err := tx.PreparexContext(ctx, sql)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, merr := stmt.ExecContext(ctx,
		userId,
	)
	return merr
}

func (p psqlRepository) insertPermissions(ctx context.Context, tx *sqlx.Tx, userId uuid.UUID, permissions []uuid.UUID) error {
	sql := `
		INSERT INTO user_permissions (user_id, permission_id)
		VALUES (?, ?)
	`

	sql = sqlx.Rebind(sqlx.DOLLAR, sql)
	stmt, err := tx.PreparexContext(ctx, sql)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, permissionId := range permissions {
		if _, err := stmt.ExecContext(ctx, userId, permissionId); err != nil {
			return err
		}
	}
	return nil
}

// Save persists a user to storage. If the user has no ID set (ID == 0),
// the implementation auto-generates one.
func (p psqlRepository) Save(ctx context.Context, user *models.User) error {
	var tx, err = p.client.GetClient().Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := p.upsertUser(ctx, tx, user); err != nil {
		return err
	}
	if err := p.deleteUserPermission(ctx, tx, user.ID); err != nil {
		return err
	}
	if err := p.insertPermissions(ctx, tx, user.ID, models.Permissions(user.Permissions.Data()).GetIds()); err != nil {
		return err
	}

	return tx.Commit()
}

// Delete removes a user by their ID. Returns an error if the user does not exist.
func (p psqlRepository) Delete(ctx context.Context, id uuid.UUID) error {
	var tx, err = p.client.GetClient().Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	sql := `
	UPDATE users
	SET deleted_at = NOW()
	WHERE id = ?
	`

	sql = sqlx.Rebind(sqlx.DOLLAR, sql)
	stmt, err := tx.PreparexContext(ctx, sql)
	if err != nil {
		return err
	}
	defer stmt.Close()

	if _, merr := stmt.ExecContext(ctx, id); merr != nil {
		return merr
	}

	return tx.Commit()
}

func (p psqlRepository) whereConds(params ListParams) ([]string, []any) {
	var conds = []string{}
	var vals = []any{}

	if v := params.Searchword; v != "" {
		sql := `users.email LIKE CONCAT('%', ?::text, '%')`
		conds = append(conds, sql)
		vals = append(vals, v)
	}

	if v := params.Role; v.Valid() {
		sql := `permissions.name = ?`
		conds = append(conds, sql)
		vals = append(vals, string(params.Role))
	}

	return conds, vals
}

// List returns all users currently stored. Never returns nil, may return an empty slice.
func (p psqlRepository) List(ctx context.Context, params ListParams) (models.Paginator[[]*models.User], error) {
	var paginator = models.Paginator[[]*models.User]{Page: params.Page, PageSize: params.PageSize}
	var where string
	var conds, vals = p.whereConds(params)
	if len(conds) > 0 {
		where = " AND " + strings.Join(conds, " AND ")
	}

	templateCountSQL := fmt.Sprintf(`
		SELECT 
			COUNT(DISTINCT users.id) as total_documents
		FROM 
			users
		LEFT JOIN
			user_permissions
		ON
			users.id = user_permissions.user_id
		LEFT JOIN
			permissions
		ON
			permissions.id = user_permissions.permission_id
		WHERE 
			users.deleted_at IS NULL
			%s
	`,
		where,
	)

	templateDataSQL := fmt.Sprintf(`
		SELECT 
			users.*,
			COALESCE(
				json_agg(row_to_json(permissions)) FILTER (WHERE permissions.id IS NOT NULL),
				'[]'::json
			) AS permissions
		FROM 
			users
		LEFT JOIN
			user_permissions
		ON
			users.id = user_permissions.user_id
		LEFT JOIN
			permissions
		ON
			permissions.id = user_permissions.permission_id
		WHERE 
			users.deleted_at IS NULL
			%s
		GROUP BY users.id
		LIMIT %d
		OFFSET %d
	`,
		where,
		paginator.Limit(),
		paginator.Offset(),
	)

	var execute = func(ctx context.Context, query string) (models.Paginator[[]*models.User], error) {
		query = sqlx.Rebind(sqlx.DOLLAR, query)
		stmt, err := p.client.GetClient().PreparexContext(ctx, query)
		if err != nil {
			return paginator, err
		}
		defer stmt.Close()

		rows, err := stmt.Unsafe().QueryxContext(ctx, vals...)
		if err != nil {
			return paginator, err
		}
		defer rows.Close()

		users, err := p.orm(ctx, rows, &paginator)
		if err != nil {
			return paginator, nil
		}

		paginator.SetDocument(users)
		return paginator, nil
	}

	pagiantor, err := execute(ctx, templateDataSQL)
	if err != nil {
		return pagiantor, err
	}
	totalDocuments, err := execute(ctx, templateCountSQL)
	if err != nil {
		return pagiantor, err
	}

	pagiantor.SetTotalDocument(totalDocuments.TotalDocument)
	return pagiantor, nil
}

func (p psqlRepository) orm(ctx context.Context, rows *sqlx.Rows, paginator *models.Paginator[[]*models.User]) ([]*models.User, error) {
	var ptrs = make([]*models.User, 0)
	type tmp struct {
		TotalDocument zero.Int64 `db:"total_documents"`
		models.User
	}

	for rows.Next() {
		var ptr = tmp{}
		if err := rows.StructScan(&ptr); err != nil {
			return nil, err
		}

		if paginator != nil {
			paginator.SetTotalDocument(ptr.TotalDocument.ValueOrZero())
		}

		ptrs = append(ptrs, &ptr.User)
	}

	if paginator != nil {
		paginator.SetDocument(paginator.Documents)
	}

	return ptrs, nil
}
