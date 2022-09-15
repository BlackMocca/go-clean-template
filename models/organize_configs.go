package models

import (
	"github.com/gofrs/uuid"
)

type OrganizesConfig struct {
	TableName   struct{}   `json:"-" db:"organize_configs" pk:"OrganizeId,ConfigKey"`
	OrganizeId  *uuid.UUID `json:"organize_id" db:"organize_id" type:"uuid"`
	ConfigKey   string     `json:"config_key" db:"config_key" type:"string"`
	ConfigValue string     `json:"config_value" db:"config_value" type:"string"`
}

type OrganizeConfig []*OrganizesConfig
