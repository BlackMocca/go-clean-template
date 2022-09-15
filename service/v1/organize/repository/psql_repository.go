package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"reflect"

	helperModel "git.innovasive.co.th/backend/models"
	"git.innovasive.co.th/backend/psql"
	"github.com/Blackmocca/go-clean-template/constants"
	"github.com/Blackmocca/go-clean-template/models"
	"github.com/Blackmocca/go-clean-template/orm"
	"github.com/Blackmocca/go-clean-template/service/v1/organize"
	"github.com/fatih/structs"
	"github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/cast"
)

type psqlOrganizeRepository struct {
	client *psql.Client
}

func NewPsqlOrganizeRepository(client *psql.Client) organize.OrganizeRepository {
	return &psqlOrganizeRepository{
		client: client,
	}
}

func (p psqlOrganizeRepository) FetchAll(ctx context.Context, args *sync.Map, paginator *helperModel.Paginator) ([]*models.Organize, error) {
	var paginateSQL string

	if paginator != nil {
		limit := paginator.PerPage
		skip := int(paginator.PerPage * (paginator.Page - 1))
		paginateSQL = fmt.Sprintf(`
		LIMIT %d
		OFFSET %d
		`, limit, skip)
	}

	var searchWord string
	if v, ok := args.Load("search_word"); ok {
		if v != nil {
			searchWord = strings.ToLower(v.(string))
		}
	}

	searchWord = fmt.Sprintf("%%%s%%", searchWord)

	sql := fmt.Sprintf(
		`
		SELECT
			organizes.total_row,
			%s, 
			%s
		FROM
			(
				SELECT
					*,
					COUNT(*) OVER() as "total_row"
				FROM
					organizes
				WHERE
					organizes.deleted_at IS NULL
				%s
			) as organizes
		JOIN
			organize_configs
		ON
			organizes.id = organize_configs.organize_id 
		WHERE
			(
				organizes.id::text=$1
			OR
				LOWER(organizes.name) like $1
			OR
				LOWER(organizes.alias_name) like $1
			)
			AND
				organizes.deleted_at IS NULL	
		ORDER BY
			organizes.order_no ASC
	`,
		orm.GetSelector(models.Organize{}),
		orm.GetSelector(models.OrganizesConfig{}),
		paginateSQL,
	)

	stmt, err := p.client.GetClient().PreparexContext(ctx, sql)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.QueryxContext(ctx, searchWord)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	relations := []string{models.FIELD_FK_ORGANIZE_CONFIG}
	org, err := p.orm(ctx, rows, relations, paginator)
	if err != nil {
		return nil, err
	}
	return org, err
}

func (p psqlOrganizeRepository) FetchOneById(ctx context.Context, orgId *uuid.UUID) (*models.Organize, error) {

	sql := fmt.Sprintf(
		`
		SELECT
			%s,
			%s
		FROM
			organizes
		LEFT JOIN
			organize_configs
		ON
			organizes.id = organize_configs.organize_id
		WHERE 
			id::text=$1
	`,
		orm.GetSelector(models.Organize{}),
		orm.GetSelector(models.OrganizesConfig{}),
	)
	stmt, err := p.client.GetClient().PreparexContext(ctx, sql)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Queryx(orgId.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	relations := []string{models.FIELD_FK_ORGANIZE_CONFIG}

	orgs, err := p.orm(ctx, rows, relations, nil)
	if err != nil {
		return nil, err
	}

	if len(orgs) > 0 {
		return orgs[0], nil
	}

	return nil, nil
}

//*Create Organizes
func (p psqlOrganizeRepository) Create(ctx context.Context, organize *models.Organize) error {
	tx, err := p.client.GetClient().Beginx()
	if err != nil {
		return err
	}

	if err := p.createOrg(ctx, tx, organize); err != nil {
		return err
	}

	if err := p.createOrgConfig(ctx, tx, organize.Id, organize.GetOrganizeConfig()); err != nil {
		return err
	}

	return tx.Commit()
}

func (p psqlOrganizeRepository) createOrg(ctx context.Context, tx *sqlx.Tx, Organize *models.Organize) error {
	sql := `
		INSERT INTO organizes (id, name, alias_name, org_type, created_at, updated_at, order_no)
		VALUES(
			$1::uuid,
			$2::text,
			$3::text,
			$4::ORG_TYPE,
			$5::timestamp,
			$6::timestamp,
			$7::numeric
		)
	`
	stmt, err := tx.Preparex(sql)
	if err != nil {
		return err
	}
	defer stmt.Close()

	if _, err := stmt.ExecContext(ctx,
		Organize.Id,
		Organize.Name,
		Organize.AliasName,
		Organize.OrgType,
		Organize.CreatedAt,
		Organize.UpdatedAt,
		Organize.OrderNo,
	); err != nil {
		if strings.Contains(err.Error(), constants.ERROR_UNIQUE_ORGANIZE_NAME) {
			return errors.New(constants.ERROR_ORGANIZE_NAME_WAS_DUPLICATE)
		}

		if strings.Contains(err.Error(), constants.ERROR_UNIQUE_ORGANIZE_ALIAS_NAME) {
			return errors.New(constants.ERROR_ORGANIZE_ALIAS_NAME_WAS_DUPLICATE)
		}

		return err
	}

	return nil
}

func (p psqlOrganizeRepository) createOrgConfig(ctx context.Context, tx *sqlx.Tx, organizeId *uuid.UUID, organizesConfig []*models.OrganizesConfig) error {
	sql := `INSERT INTO organize_configs(organize_id, config_key, config_value)
	VALUES (
		$1::uuid,
		$2::text,
		$3::text
	)
	`

	stmt, err := tx.Preparex(sql)
	if err != nil {
		return err
	}
	defer stmt.Close()

	if len(organizesConfig) > 0 {
		for _, config := range organizesConfig {
			if _, err := stmt.ExecContext(ctx, organizeId, config.ConfigKey, config.ConfigValue); err != nil {
				return err
			}
		}
	}

	return nil
}

func (p psqlOrganizeRepository) bindingConfigValue(ptr interface{}, config []*models.OrganizesConfig) {
	if config != nil && len(config) > 0 {
		var faith = structs.New(ptr)
		var fields = faith.Fields()
		for _, c := range config {
			key := c.ConfigKey
			val := c.ConfigValue

			if val != "" {
				for fieldIndex := range fields {
					configKey := fields[fieldIndex].Tag("config_key")
					if configKey != "" && strings.Contains(key, configKey) {
						fieldType := reflect.TypeOf(fields[fieldIndex].Value()).Kind()
						switch fieldType {
						case reflect.Slice:
							currentSlice := reflect.ValueOf(fields[fieldIndex].Value()).Interface().([]string)
							currentSlice = append(currentSlice, val)
							fields[fieldIndex].Set(currentSlice)
						case reflect.String:
							fields[fieldIndex].Set(val)
						case reflect.Int64:
							fields[fieldIndex].Set(cast.ToInt64(val))
						case reflect.Bool:
							fields[fieldIndex].Set(cast.ToBool(val))
						default:
							uid := uuid.FromStringOrNil(val)
							fields[fieldIndex].Set(&uid)
						}

					}
				}
			}
		}
	}
}

func (p psqlOrganizeRepository) orm(ctx context.Context, rows *sqlx.Rows, relationships []string, paginator *helperModel.Paginator) ([]*models.Organize, error) {
	var ptrs = make([]*models.Organize, 0)
	var mapper, err = orm.NewRowsScan(rows)
	if err != nil {
		return nil, err
	}

	if mapper.TotalRows() > 0 {
		for _, row := range mapper.RowsValues() {
			var ptr = new(models.Organize)

			ptr, err := orm.OrmOrganize(ptr, mapper, row, relationships)
			if err != nil {
				return nil, err
			}
			if ptr != nil {
				exists, err := orm.IsDuplicateByPK(ptrs, ptr)
				if err != nil {
					return nil, err
				}
				if !exists {
					ptrs = append(ptrs, ptr)
				}

			}
		}

	}

	if paginator != nil {
		paginator.SetPaginatorByAllRows(mapper.PaginateTotal())
	}

	if len(ptrs) > 0 {
		for index, _ := range ptrs {
			if err := orm.OrmOrganizeRelation(ptrs[index], mapper, relationships); err != nil {
				return nil, err
			}
			p.bindingConfigValue(ptrs[index], ptrs[index].Configs)
		}
	}

	return ptrs, nil
}
