package mysql_base

import (
	"strings"
)

type FieldConfig struct {
	Name          string `json:"name"`
	Type          string `json:"type"`
	Length        int    `json:"length"`
	IndexType     string `json:"index_type"`
	CreateFlags   string `json:"create_flags"`
	RealType      int
	RealIndexType int
}

type TableConfig struct {
	Name       string         `json:"name"`
	PrimaryKey string         `json:"primary_key"`
	Engine     string         `json:"engine"`
	Fields     []*FieldConfig `json:"fields"`
}

func (this *TableConfig) GetPrimaryKeyFieldConfig() (field_config *FieldConfig) {
	if this.PrimaryKey == "" || this.Fields == nil {
		return nil
	}

	for _, f := range this.Fields {
		if f.Name == this.PrimaryKey {
			field_config = f
			break
		}
	}

	return
}

func (this *TableConfig) IsPrimaryAutoIncrement() bool {
	f := this.GetPrimaryKeyFieldConfig()
	if f == nil {
		return false
	}
	strs := strings.Split(f.CreateFlags, ",")
	for _, s := range strs {
		if c, o := GetMysqlTableCreateFlagTypeByString(strings.ToUpper(s)); o {
			if c == MYSQL_TABLE_CREATE_AUTOINCREMENT {
				return true
			}
		}
	}
	return false
}

func (this *TableConfig) GetField(field_name string) *FieldConfig {
	var field *FieldConfig
	if this.Fields != nil {
		for _, f := range this.Fields {
			if f.Name == field_name {
				field = f
				break
			}
		}
	}
	return field
}
