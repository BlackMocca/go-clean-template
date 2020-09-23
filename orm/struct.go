package orm

import (
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
)

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
		valInt64, err := strconv.ParseInt(value, 10, 64)
		if err == nil {
			field.Set(valInt64)
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
	}
	return nil
}

/*
	Equal Value if a same type
*/
func equal(typeVal string, x interface{}, y interface{}) bool {
	if x != nil && y != nil {
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
