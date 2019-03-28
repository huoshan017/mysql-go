package mysql_base

type FieldConfig struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Length      int    `json:"length"`
	IndexType   string `json:"index_type"`
	CreateFlags string `json:"create_flags"`
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
