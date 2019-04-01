package mysql_base

type FieldValuePair struct {
	Name  string
	Value interface{}
}

func (this *Database) InsertRecord(table_name string, field_args ...FieldValuePair) (err error) {
	return
}

func (this *Database) SelectRecord(table_name, key_field FieldValuePair) (err error) {
	return
}

func (this *Database) UpdateRecord(table_name string, key_field FieldValuePair, field_args ...FieldValuePair) (err error) {
	return
}

func (this *Database) DeleteRecord(table_name string, key_field FieldValuePair) (err error) {
	return
}
