package models

import (
	"reflect"
	"time"

	"git.innovasive.co.th/backend/helper"
	"git.innovasive.co.th/backend/models"
	"github.com/gofrs/uuid"
	"github.com/spf13/cast"
)

const UserSelector = `
		users.id			"users.id",
		users.email			"users.email",
		users.firstname		"users.firstname",
		users.lastname 		"users.lastname",
		users.age			"users.age",
		users.created_at	"users.created_at",
		users.updated_at 	"users.updated_at",
		users.deleted_at	"users.deleted_at"
`

type User struct {
	TableName struct{}          `json:"-" db:"users"`
	Id        *uuid.UUID        `json:"id" db:"id" type:"uuid"`
	Email     string            `json:"email" db:"email" type:"string"`
	Firstname string            `json:"firstname" db:"firstname" type:"string"`
	Lastname  string            `json:"lastname" db:"lastname" type:"string"`
	Age       int               `json:"age" db:"age" type:"int32"`
	CreatedAt *models.Timestamp `json:"created_at" db:"created_at" type:"timestamp"`
	UpdatedAt *models.Timestamp `json:"updated_at" db:"updated_at" type:"timestamp"`
	DeletedAt *models.Timestamp `json:"deleted_at" db:"deleted_at" type:"timestamp"`
}

type Users []*User

func NewUserWithParams(params map[string]interface{}, ptr *User) *User {
	if ptr == nil {
		ptr = new(User)
	}

	for key, v := range params {
		switch key {
		case "id":
			ptr.Id, _ = helper.ConvertToUUIDAndBinary(v)
		case "email":
			ptr.Email = cast.ToString(v)
		case "firstname":
			ptr.Firstname = cast.ToString(v)
		case "lastname":
			ptr.Lastname = cast.ToString(v)
		case "age":
			ptr.Age = cast.ToInt(v)
		case "created_at":
			if reflect.ValueOf(v).Kind() == reflect.String {
				ht := models.NewTimestampFromString(cast.ToString(v))
				ptr.CreatedAt = &ht
			} else if reflect.ValueOf(v).String() == "time.Time" {
				ht := models.NewTimestampFromTime(v.(time.Time))
				ptr.CreatedAt = &ht
			}
		case "updated_at":
			if reflect.ValueOf(v).Kind() == reflect.String {
				ht := models.NewTimestampFromString(cast.ToString(v))
				ptr.CreatedAt = &ht
			} else if reflect.ValueOf(v).String() == "time.Time" {
				ht := models.NewTimestampFromTime(v.(time.Time))
				ptr.CreatedAt = &ht
			}
		case "deleted_at":
			if reflect.ValueOf(v).Kind() == reflect.String {
				ht := models.NewTimestampFromString(cast.ToString(v))
				ptr.CreatedAt = &ht
			} else if reflect.ValueOf(v).String() == "time.Time" {
				ht := models.NewTimestampFromTime(v.(time.Time))
				ptr.CreatedAt = &ht
			}

		}
	}

	return ptr
}

func (u *User) GenUUID() {
	uid, _ := uuid.NewV4()
	u.Id = &uid
}
