package models

import (
	"git.innovasive.co.th/backend/models"
	"github.com/gofrs/uuid"
)

type UserType struct {
	TableName struct{}          `json:"-" db:"user_types" pk:"Id"`
	Id        *uuid.UUID        `json:"id" db:"id" type:"uuid"`
	Name      string            `json:"name" db:"name" type:"string"`
	CreatedAt *models.Timestamp `json:"created_at" db:"created_at" type:"timestamp"`
	UpdatedAt *models.Timestamp `json:"updated_at" db:"updated_at" type:"timestamp"`
}
