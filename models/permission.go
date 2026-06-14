package models

import (
	utilModels "github.com/BlackMocca/utils/models"
	"github.com/google/uuid"
	"github.com/guregu/null/v6/zero"
)

type Permission struct {
	Id        uuid.UUID             `json:"id" db:"id"`
	Name      zero.String           `json:"name" db:"name"`
	CreatedAt *utilModels.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt *utilModels.Timestamp `json:"updated_at" db:"updated_at"`
	DeletedAt *utilModels.Timestamp `json:"deleted_at" db:"deleted_at"`
}

type Permissions []*Permission

func (p Permissions) GetIds() []uuid.UUID {
	var ptrs = make([]uuid.UUID, 0, len(p))
	for _, item := range p {
		ptrs = append(ptrs, item.Id)
	}
	return ptrs
}
