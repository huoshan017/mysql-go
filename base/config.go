package mysql_base

import (
	"strings"
)

type FieldConfig struct {
	Name       string `json:"name"`
	TypeStr    string `json:"type"`
	IndexStr   string `json:"index"`
	Type       int
	Index      int
	StructName string
}

type TableConfig struct {
	Name       string         `json:"name"`
	PrimaryKey string         `json:"primary_key"`
	SingleRow  bool           `json:"single_row"`
	Engine     string         `json:"engine"`
	Fields     []*FieldConfig `json:"fields"`
	FieldMap   map[string]*FieldConfig
}

func (tc *TableConfig) AfterLoad() {
	for _, f := range tc.Fields {
		if tc.FieldMap == nil {
			tc.FieldMap = make(map[string]*FieldConfig)
		}
		tc.FieldMap[f.Name] = f
	}
}

func (tc *TableConfig) GetField(field_name string) *FieldConfig {
	if tc.FieldMap == nil {
		return nil
	}
	return tc.FieldMap[field_name]
}

func (tc *TableConfig) GetPrimaryKeyFieldConfig() (field_config *FieldConfig) {
	if tc.PrimaryKey == "" || tc.Fields == nil {
		return nil
	}

	for _, f := range tc.Fields {
		if f.Name == tc.PrimaryKey {
			field_config = f
			break
		}
	}

	return
}

func (tc *TableConfig) IsPrimaryAutoIncrement() bool {
	f := tc.GetPrimaryKeyFieldConfig()
	if f == nil {
		return false
	}

	if !strings.Contains(f.TypeStr, "AUTO_INCREMENT") {
		return false
	}

	return true
}

func (tc *TableConfig) HasBytesField() bool {
	if tc.Fields == nil {
		return false
	}
	for _, f := range tc.Fields {
		if IsMysqlFieldBinaryType(f.Type) || IsMysqlFieldBlobType(f.Type) {
			return true
		}
	}
	return false
}

func (tc *TableConfig) HasStructField() bool {
	if tc.Fields == nil {
		return false
	}
	for _, f := range tc.Fields {
		if f.StructName != "" {
			return true
		}
	}
	return false
}
