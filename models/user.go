// Package models defines domain types shared across the example application.
package models

import (
	"example/constants"

	utilModels "github.com/BlackMocca/utils/models"
	"github.com/google/uuid"
	"github.com/guregu/null/v6/zero"
)

// User represents a user entity in the system.
type User struct {
	ID uuid.UUID `json:"id" db:"id"`
	// Name is the display name of the user.
	Name zero.String `json:"name" db:"name"`
	// Email is the user's contact email address.
	Email     zero.String                         `json:"email" db:"email"`
	Role      utilModels.EnumScan[constants.Role] `json:"role" db:"role"`
	CreatedAt *utilModels.Timestamp               `json:"created_at" db:"created_at"`
	UpdatedAt *utilModels.Timestamp               `json:"updated_at" db:"updated_at"`
	DeletedAt *utilModels.Timestamp               `json:"deleted_at" db:"deleted_at"`

	Permissions utilModels.JsonScan[[]*Permission] `json:"permissions" db:"permissions"`
}
