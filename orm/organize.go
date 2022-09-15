package orm

import (
	"errors"

	"github.com/Blackmocca/go-clean-template/models"
	"github.com/fatih/structs"
)

func OrmOrganize(ptr *models.Organize, mapper RowScan, currentRow RowValue, relationFieldNames []string) (*models.Organize, error) {
	v, err := fillValue(ptr, currentRow)
	if v != nil {
		return v.(*models.Organize), nil
	}

	return nil, err
}

func OrmOrganizeRelation(ptr *models.Organize, mapper RowScan, relationFieldNames []string) error {
	if ptr != nil && relationFieldNames != nil {
		if len(relationFieldNames) > 0 {
			faithStruct, _ := GetStructFields(ptr)
			for _, fieldName := range relationFieldNames {
				f := faithStruct.Field(fieldName)
				switch f.Name() {
				case models.FIELD_FK_ORGANIZE_CONFIG:
					fkVal := getFKTag(f.Tag("fk"))
					fkRelation, ok := fkVal.Load("relation")
					if !ok {
						return errors.New(ERROR_NO_FIND_RELATION)
					}
					var fkPtrs = make([]*models.OrganizesConfig, 0)

					for _, rowValue := range mapper.RowsValues() {
						var linkPtr = new(models.OrganizesConfig)
						var err error
						var primaryKeys = GetPK(linkPtr)
						if len(primaryKeys) == 0 {
							return errors.New(ERROR_NO_FIND_PK)
						}
						linkPtr, err = OrmOraganizeConfig(linkPtr, mapper, rowValue, relationFieldNames)
						if err != nil {
							return err
						}

						if linkPtr != nil {
							var fkCol1 interface{}
							var fkCol2 interface{}
							fkCol1, _ = fkVal.Load("fk_field1")
							fkCol2, _ = fkVal.Load("fk_field2")
							if fkCol1 == nil || fkCol2 == nil {
								return errors.New("fk tag: fk_field1 or fk_field2 not found")
							}

							parentID := faithStruct.Field(fkCol1.(string)).Value()
							parentType := faithStruct.Field(fkCol1.(string)).Tag("type")
							linkID := structs.New(linkPtr).Field(fkCol2.(string)).Value()
							if equal(parentType, parentID, linkID) {
								exists, err := IsDuplicateByPK(fkPtrs, linkPtr)
								if err != nil {
									return err
								}
								if !exists {
									fkPtrs = append(fkPtrs, linkPtr)
								}
							}
						}
					}

					switch fkRelation {
					case relationONE:
						f.Set(fkPtrs[0])
					default:
						f.Set(fkPtrs)
					}

				}
			}
		}
	}

	return nil
}
