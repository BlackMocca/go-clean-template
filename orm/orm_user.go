package orm

import (
	"strings"

	"github.com/BlackMocca/go-clean-template/models"
	"github.com/fatih/structs"
	"github.com/jmoiron/sqlx"
)

func OrmUser(model *models.User, rows *sqlx.Rows, relationFieldNames []string) (*models.User, error) {
	orgTableName := GetTableName(model)

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	_, ptrColumnMap := GetStructFields(model)

	if err := rows.Err(); err != nil {
		return nil, err
	}

	values, err := rows.SliceScan()
	if err != nil {
		return nil, err
	}

	if len(values) > 0 {
		for index, col := range columns {
			orderCol := strings.ReplaceAll(col, orgTableName+".", "")
			if field, ok := ptrColumnMap.Load(orderCol); ok {
				if err := SetFieldFromType(field.(*structs.Field), values[index]); err != nil {
					return nil, err
				}
			}
		}
	}

	if model.IsZero() {
		return nil, nil
	}

	return model, nil
}
