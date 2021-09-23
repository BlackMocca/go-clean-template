package orm

import (
	"github.com/BlackMocca/go-clean-template/models"
)

func OrmUserType(ptr *models.UserType, mapper RowScan, currentRow RowValue, relationFieldNames []string) (*models.UserType, error) {
	v, err := fillValue(ptr, currentRow)
	if v != nil {
		return v.(*models.UserType), nil
	}

	return nil, err
}
