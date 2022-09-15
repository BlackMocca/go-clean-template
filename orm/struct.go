package orm

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"git.innovasive.co.th/backend/models"
	"github.com/fatih/structs"
	"github.com/gofrs/uuid"
	"github.com/guregu/null/zero"
	"github.com/spf13/cast"
)

const (
	typeUUID       = "uuid"
	typeZeroUUID   = "zerouuid"
	typeString     = "string"
	typeInt32      = "int32"
	typeInt64      = "int64"
	typeFloat64    = "float64"
	typeTimeStamp  = "timestamp"
	typeDate       = "date"
	typeZeroString = "zerostring"
	typeZeroInt    = "zeroint"
	typeZeroFloat  = "zerofloat"
	typeZeroBool   = "zerobool"
	typeDuration   = "duration"
	typeBool       = "bool"

	relationONE  = "one"
	relationMANY = "many"

	ERROR_NO_FIND_PK       = "can not file pk tag on field TableName"
	ERROR_NO_FIND_RELATION = "can not find relation on tag fk"
)

var (
	patternSelector = func(tablename string, fieldDB string) string {
		return fmt.Sprintf(`%s.%s "%s.%s"`, tablename, fieldDB, tablename, fieldDB)
	}
)

func IsDuplicateByPK(modelsSlice interface{}, model interface{}) (bool, error) {
	sliceMapItem := make([]reflect.Value, 0)
	if reflect.TypeOf(modelsSlice).Kind() == reflect.Slice {
		if reflect.ValueOf(modelsSlice).Len() > 0 {
			for i := 0; i < reflect.ValueOf(modelsSlice).Len(); i++ {
				m := reflect.ValueOf(modelsSlice).Index(i).Elem()
				sliceMapItem = append(sliceMapItem, m)
			}
		}
	}

	if len(sliceMapItem) > 0 {
		var fieldPrimaryKey = GetPK(model)
		var existsPKFieldAmount = len(fieldPrimaryKey)
		if existsPKFieldAmount == 0 {
			return false, errors.New(ERROR_NO_FIND_PK)
		}

		for _, value := range sliceMapItem {
			var checkPKAmount int
			for _, pkField := range fieldPrimaryKey {
				field := value.FieldByName(pkField)
				var pkDataInSlice string
				if field.Kind() == reflect.Ptr || field.Kind() == reflect.Interface {
					pkDataInSlice = cast.ToString(field.Elem().Interface())
				} else {
					pkDataInSlice = cast.ToString(field.Interface())
				}

				var newmodelPKData string
				var modelfield = reflect.ValueOf(model)
				if modelfield.Kind() == reflect.Ptr || modelfield.Kind() == reflect.Interface {
					newmodelPKData = cast.ToString(modelfield.Elem().FieldByName(pkField).Interface())
				} else {
					newmodelPKData = cast.ToString(modelfield.FieldByName(pkField).Interface())
				}
				if pkDataInSlice == newmodelPKData {
					checkPKAmount++
				}
			}
			if checkPKAmount == existsPKFieldAmount {
				return true, nil
			}
		}
	}
	return false, nil
}

func GetValueFromTag(model interface{}, field string, tag string) string {
	var tagVal string

	m := structs.New(model)
	if f, ok := m.FieldOk(field); ok {
		tagVal = f.Tag(tag)
	}

	return tagVal
}

func GetStructFields(models interface{}) (*structs.Struct, *sync.Map) {
	var ptrColumnMap = new(sync.Map)
	faithOrder := structs.New(models)
	fields := faithOrder.Fields()
	for _, f := range fields {
		tagCol := f.Tag("db")
		if tagCol != "" && tagCol != "-" {
			ptrColumnMap.Store(tagCol, f)
		}
	}

	return faithOrder, ptrColumnMap
}

func GetTableName(models interface{}) string {
	var tablename string
	faith := structs.New(models)

	if f, ok := faith.FieldOk("TableName"); ok {
		tablename = f.Tag("db")
	}

	return tablename
}

func GetPK(models interface{}) []string {
	var pks = make([]string, 0)
	faith := structs.New(models)

	if f, ok := faith.FieldOk("TableName"); ok {
		pks = strings.Split(f.Tag("pk"), ",")
	}

	return pks
}

func GetSelector(models interface{}) string {
	faith := structs.New(models)
	fields := faith.Fields()
	tablename := GetTableName(models)
	var selectors = make([]string, 0)

	if len(fields) > 0 {
		for _, field := range fields {
			if field.Name() != "TableName" {
				columename := field.Tag("db")
				if columename != "" && columename != "-" {
					selectors = append(selectors, patternSelector(tablename, columename))
				}
			}
		}
	}

	return strings.Join(selectors, ",")
}

func SetFieldFromType(field *structs.Field, v interface{}) error {
	var value string
	var tag = field.Tag("type")
	if v != nil {
		if reflect.TypeOf(v).String() == "time.Time" {
			switch tag {
			case typeDate:
				value = v.(time.Time).Format(models.DateLayout)
			case typeTimeStamp:
				value = v.(time.Time).Format(models.TimestampLayout)
			}

		} else {
			value = cast.ToString(v)
		}
	}

	switch tag {
	case typeUUID:
		uid, err := uuid.FromString(value)
		if err == nil {
			field.Set(&uid)
		}
	case typeZeroUUID:
		uid, err := models.NewZeroUUIDFromstring(value)
		if err == nil {
			field.Set(uid)
		}
	case typeString:
		field.Set(value)
	case typeInt32:
		valInt, err := strconv.Atoi(value)
		if err == nil {
			field.Set(valInt)
		}
	case typeInt64:
		valInt, err := strconv.ParseInt(value, 10, 64)
		if err == nil {
			field.Set(valInt)
		}
	case typeFloat64:
		f, err := strconv.ParseFloat(value, 64)
		if err == nil {
			field.Set(f)
		}
	case typeTimeStamp:
		if value != "" {
			timestamp := models.NewTimestampFromString(value)
			field.Set(&timestamp)
		}
	case typeDate:
		if value != "" {
			date := models.NewDateFromString(value)
			field.Set(&date)
		}
	case typeZeroString:
		zeroString := zero.StringFrom(value)
		field.Set(zeroString)
	case typeZeroInt:
		var zeroInt zero.Int
		if value != "" {
			valueInt, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return err
			}
			zeroInt = zero.IntFrom(valueInt)
		} else {
			zeroInt = zero.IntFrom(0)
		}
		field.Set(zeroInt)
	case typeZeroFloat:
		valFloat, _ := strconv.ParseFloat(value, 64)
		zeroFloat := zero.FloatFrom(valFloat)
		field.Set(zeroFloat)
	case typeZeroBool:
		b, _ := strconv.ParseBool(value)
		field.Set(zero.BoolFrom(b))
	case typeDuration:
		duration := fmt.Sprintf("%sh", value)
		d, err := time.ParseDuration(duration)
		if err != nil {
			return err
		}
		field.Set(d)

	case typeBool:
		val, _ := strconv.ParseBool(value)
		field.Set(val)
	}
	return nil
}

func isNil(val interface{}) bool {
	if val == nil || (reflect.ValueOf(val).Kind() == reflect.Ptr && reflect.ValueOf(val).IsNil()) {
		return true
	}
	return false
}

/*
	Equal Value if a same type
*/
func equal(typeVal string, x interface{}, y interface{}) bool {
	if !isNil(x) && !isNil(y) {
		switch typeVal {
		case typeUUID:
			return x.(*uuid.UUID).String() == y.(*uuid.UUID).String()
		case typeString:
			if x.(string) == "" || y.(string) == "" {
				return false
			}
			return x.(string) == y.(string)
		case typeZeroUUID:
			if reflect.TypeOf(x).String() == "models.ZeroUUID" && reflect.TypeOf(y).String() == "models.ZeroUUID" {
				if x.(models.ZeroUUID) == (models.ZeroUUID{}) || y.(models.ZeroUUID) == (models.ZeroUUID{}) {
					return false
				}
				return x.(models.ZeroUUID).String() == y.(models.ZeroUUID).String()
			} else if reflect.TypeOf(x).String() == "models.ZeroUUID" && reflect.TypeOf(y).String() == "*uuid.UUID" {
				if x.(models.ZeroUUID) == (models.ZeroUUID{}) || y == nil {
					return false
				}
				return x.(models.ZeroUUID).String() == y.(*uuid.UUID).String()
			} else if reflect.TypeOf(x).String() == "*uuid.UUID" && reflect.TypeOf(y).String() == "models.ZeroUUID" {
				if x == nil || y.(models.ZeroUUID) == (models.ZeroUUID{}) {
					return false
				}
				return y.(models.ZeroUUID).String() == x.(*uuid.UUID).String()
			}
		}
	}
	return false
}

func getFKTag(tag string) *sync.Map {
	var m = sync.Map{}
	if tag == "" {
		return &m
	}

	vals := strings.Split(tag, ",")
	for _, val := range vals {
		keyVal := strings.Split(val, ":")
		key := keyVal[0]
		value := keyVal[1]
		m.Store(key, value)
	}

	return &m
}

func fillValue(ptr interface{}, currentRow RowValue) (interface{}, error) {
	schTableName := GetTableName(ptr)

	columns := currentRow.Columns()

	_, ptrColumnMap := GetStructFields(ptr)

	values := currentRow.Values()
	if len(values) > 0 {
		for index, col := range columns {
			orderCol := strings.ReplaceAll(col, schTableName+".", "")
			if field, ok := ptrColumnMap.Load(orderCol); ok {
				if err := SetFieldFromType(field.(*structs.Field), values[index]); err != nil {
					return nil, err
				}
			}
		}
	}

	return ptr, nil
}
