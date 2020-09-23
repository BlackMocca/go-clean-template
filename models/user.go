package models

import (
	"git.innovasive.co.th/backend/models"
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
	ID        int64             `json:"id" db:"id" type:"int64"`
	Email     string            `json:"email" db:"email" type:"string"`
	Firstname string            `json:"firstname" db:"firstname" type:"string"`
	Lastname  string            `json:"lastname" db:"lastname" type:"string"`
	Age       int               `json:"age" db:"age" type:"int32"`
	CreatedAt *models.Timestamp `json:"created_at" db:"created_at" type:"timestamp"`
	UpdatedAt *models.Timestamp `json:"updated_at" db:"updated_at" type:"timestamp"`
	DeletedAt *models.Timestamp `json:"deleted_at" db:"deleted_at" type:"timestamp"`
}

type Users []*User

func (u User) IsZero() bool {
	if u.DeletedAt != nil {
		return true
	}
	return false
}
