package mysql_base

import (
	"strings"
)

type FieldConfig struct {
	Name          string `json:"name"`
	Type          string `json:"type"`
	IndexType     string `json:"index_type"`
	RealType      int
	RealIndexType int
	StructName    string
}

type TableConfig struct {
	Name       string         `json:"name"`
	PrimaryKey string         `json:"primary_key"`
	SingleRow  bool           `json:"single_row"`
	Engine     string         `json:"engine"`
	Fields     []*FieldConfig `json:"fields"`
	FieldMap   map[string]*FieldConfig
}

func (this *TableConfig) AfterLoad() {
	for _, f := range this.Fields {
		if this.FieldMap == nil {
			this.FieldMap = make(map[string]*FieldConfig)
		}
		this.FieldMap[f.Name] = f
	}
}

func (this *TableConfig) GetField(field_name string) *FieldConfig {
	if this.FieldMap == nil {
		return nil
	}
	return this.FieldMap[field_name]
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

	if !strings.Contains(f.Type, "AUTO_INCREMENT") {
		return false
	}

	return true
}

func (this *TableConfig) HasBytesField() bool {
	if this.Fields == nil {
		return false
	}
	for _, f := range this.Fields {
		if IsMysqlFieldBinaryType(f.RealType) || IsMysqlFieldBlobType(f.RealType) {
			return true
		}
	}
	return false
}

func (this *TableConfig) HasStructField() bool {
	if this.Fields == nil {
		return false
	}
	for _, f := range this.Fields {
		if f.StructName != "" {
			return true
		}
	}
	return false
}
