package models

import (
	"encoding/json"
	"reflect"
	"strings"
	"time"

	"git.innovasive.co.th/backend/helper"
	helperModel "git.innovasive.co.th/backend/models"
	"github.com/fatih/structs"
	"github.com/gofrs/uuid"
	"github.com/guregu/null/zero"
	"github.com/spf13/cast"
)

const (
	FIELD_FK_ORGANIZE_CONFIG = "Configs"

	tag_config_key = "config_key"
	tag_type_key   = "type"
)

type Organize struct {
	TableName struct{}    `json:"-" db:"organizes" pk:"Id"`
	Id        *uuid.UUID  `json:"id" db:"id" type:"uuid"`
	Name      string      `json:"name" db:"name" type:"string"`
	AliasName zero.String `json:"alias_name" db:"alias_name" type:"zerostring"`
	OrgType   string      `json:"org_type" db:"org_type" type:"string"`
	OrderNo   int64       `json:"order_no" db:"order_no" type:"int64"`
	Admin1    *uuid.UUID  `json:"admin_1" db:"-" config_key:"admin_1" type:"uuid"`
	Admin2    *uuid.UUID  `json:"admin2_" db:"-" config_key:"admin_2" type:"uuid"`

	Configs []*OrganizesConfig `json:"-" db:"-" fk:"relation:many,fk_field1:Id,fk_field2:OrganizeId"`

	CreatedAt *helperModel.Timestamp `json:"created_at" db:"created_at" type:"timestamp"`
	UpdatedAt *helperModel.Timestamp `json:"updated_at" db:"updated_at" type:"timestamp"`
	DeletedAt *helperModel.Timestamp `json:"deleted_at" db:"deleted_at" type:"timestamp"`
}

func (o *Organize) GetOrganizeConfig() []*OrganizesConfig {
	configs := make([]*OrganizesConfig, 0)

	s := structs.New(o)
	fields := s.Fields()

	for _, f := range fields {
		tagType := f.Tag(tag_type_key)
		if tagType == "" {
			continue
		}
		tag := f.Tag(tag_config_key)
		if tag == "" {
			continue
		}
		switch tagType {
		case "string":
			configs = append(configs, &OrganizesConfig{
				ConfigKey:   tag,
				ConfigValue: cast.ToString(f.Value()),
			})
		case "int64":
			configs = append(configs, &OrganizesConfig{
				ConfigKey:   tag,
				ConfigValue: cast.ToString(f.Value()),
			})
		case "uuid":
			if f.Value().(*uuid.UUID) == nil {
				configs = append(configs, &OrganizesConfig{
					ConfigKey:   tag,
					ConfigValue: "",
				})
			} else {
				configs = append(configs, &OrganizesConfig{
					ConfigKey:   tag,
					ConfigValue: cast.ToString(f.Value()),
				})
			}

		}
	}

	return configs
}

func NewOrganizeWithParams(params map[string]interface{}, ptr *Organize) *Organize {
	if ptr == nil {
		ptr = new(Organize)
	}

	for key, val := range params {
		switch key {
		case "id":
			ptr.Id, _ = helper.ConvertToUUIDAndBinary(val)
		case "name":
			ptr.Name = strings.ToLower(cast.ToString(val))
		case "alias_name":
			ptr.AliasName = zero.StringFrom(cast.ToString(val))
		case "org_type":
			ptr.OrgType = cast.ToString(val)
		case "order_no":
			ptr.OrderNo = cast.ToInt64(val)
		case "admin_1":
			if cast.ToString(val) != "" {
				ptr.Admin1, _ = helper.ConvertToUUIDAndBinary(val)
			}
		case "admin_2":
			if cast.ToString(val) != "" {
				ptr.Admin2, _ = helper.ConvertToUUIDAndBinary(val)
			}
		case "created_at":
			if val != nil {
				if reflect.TypeOf(val).Kind() == reflect.String {
					t := helperModel.NewTimestampFromString(cast.ToString(val))
					ptr.CreatedAt = &t
				} else if reflect.TypeOf(val).String() == "time.Time" {
					t := helperModel.NewTimestampFromTime(val.(time.Time))
					ptr.CreatedAt = &t
				}
			}
		case "deleted_at":
			if val != nil {
				if reflect.TypeOf(val).Kind() == reflect.String {
					t := helperModel.NewTimestampFromString(cast.ToString(val))
					ptr.UpdatedAt = &t
				} else if reflect.TypeOf(val).String() == "time.Time" {
					t := helperModel.NewTimestampFromTime(val.(time.Time))
					ptr.UpdatedAt = &t
				}
			}
		}

	}

	return ptr
}

func (o *Organize) String() string {
	bu, _ := json.Marshal(o)
	return string(bu)
}

func (o *Organize) NewUUID() {
	userId, _ := uuid.NewV4()
	o.Id = &userId
}

func (o *Organize) SetCreatedAt(ti helperModel.Timestamp) {
	o.CreatedAt = &ti
}

func (o *Organize) SetUpdatedAt(ti helperModel.Timestamp) {
	o.UpdatedAt = &ti
}
