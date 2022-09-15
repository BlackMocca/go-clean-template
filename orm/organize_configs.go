package orm

import (
	"github.com/Blackmocca/go-clean-template/models"
)

func OrmOraganizeConfig(ptr *models.OrganizesConfig, mapper RowScan, currentRow RowValue, relationFieldNames []string) (*models.OrganizesConfig, error) {
	v, err := fillValue(ptr, currentRow)
	if v != nil {
		return v.(*models.OrganizesConfig), nil
	}

	return nil, err
}
