package orm

import (
	"errors"

	"github.com/BlackMocca/go-clean-template/models"
	"github.com/fatih/structs"
)

func OrmUser(ptr *models.User, mapper RowScan, currentRow RowValue, relationFieldNames []string) (*models.User, error) {
	v, err := fillValue(ptr, currentRow)
	if v != nil {
		return v.(*models.User), nil
	}

	return nil, err
}

func OrmUserRelation(ptr *models.User, mapper RowScan, relationFieldNames []string) error {
	if ptr != nil && relationFieldNames != nil {
		if len(relationFieldNames) > 0 {
			faithStruct, _ := GetStructFields(ptr)
			for _, fieldName := range relationFieldNames {
				f := faithStruct.Field(fieldName)
				switch f.Name() {
				case models.FIELD_FK_USER_TYPE:
					fkVal := getFKTag(f.Tag("fk"))
					fkRelation, ok := fkVal.Load("relation")
					if !ok {
						return errors.New(ERROR_NO_FIND_RELATION)
					}
					if fkRelation == relationONE {
						var fkPtrs = make([]*models.UserType, 0)

						for _, rowValue := range mapper.RowsValues() {
							var linkPtr = new(models.UserType)
							var err error
							var primaryKeys = GetPK(linkPtr)
							if len(primaryKeys) == 0 {
								return errors.New(ERROR_NO_FIND_PK)
							}
							linkPtr, err = OrmUserType(linkPtr, mapper, rowValue, relationFieldNames)
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

						if len(fkPtrs) > 0 {
							f.Set(fkPtrs[0])
						}
					}

				}
			}
		}
	}

	return nil
}
