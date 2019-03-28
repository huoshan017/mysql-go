package mysql_base

import (
	"fmt"
	"log"
	"strings"
)

func (this *Database) LoadTable(tab *TableConfig) bool {
	primary_field := tab.GetPrimaryKeyFieldConfig()
	if primary_field == nil {
		log.Printf("Database::LoadTable %v cant get primary key field config\n", tab.Name)
		return false
	}

	// create table
	sql_str := fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%s` (`%s` %s %s, PRIMARY KEY(`%s`)) ENGINE=%s", tab.Name, tab.PrimaryKey, primary_field.Type, primary_field.CreateFlags, tab.PrimaryKey, tab.Engine)
	if !this.Exec(sql_str, nil, nil) {
		return false
	}

	// add fields
	for _, f := range tab.Fields {
		if tab.PrimaryKey == f.Name {
			continue
		}
	}

	return true
}

func _get_field_create_flags_str(create_flags string) (create_flags_str string) {
	flags := strings.Split(create_flags, ",")
	for _, s := range flags {
		s = strings.ToUpper(s)
		t, o := GetMysqlTableCreateFlagTypeByString(s)
		if !o {
			log.Printf("Create table flag %v not found\n", s)
			break
		}
		// 缺省
		if t == MYSQL_TABLE_CREATE_DEFAULT {
			if IsMysqlFieldIntType(t) {
				create_flags_str += (s + " 0")
			} else if IsMysqlFieldTextType(t) || IsMysqlFieldBinaryType(t) || IsMysqlFieldBlobType(t) {
				create_flags_str += (s + " ''")
			} else if IsMysqlFieldTimestampType(t) {
				create_flags_str += (s + " CURRENT_TIMESTAMP")
			} else {
				log.Printf("Create table flag %v no default value", s)
			}
		} else {
			create_flags_str += s
		}
	}
	return
}

func (this *Database) add_field(table_name string, field *FieldConfig) bool {
	var result QueryResultList
	sql_str := fmt.Sprintf("DESCRIBE %s %s", table_name, field.Name)
	if !this.Query(sql_str, &result) {
		return false
	}

	if result.rows != nil || result.Get() {
		log.Printf("describe get rows not empty")
		return false
	}

	create_flags_str := _get_field_create_flags_str(field.CreateFlags)

	sql_str = fmt.Sprintf("ALTER TABLE `%s` ADD COLUMN `%s` %s %s", table_name, field.Name, field.Type, create_flags_str)
	if !this.Exec(sql_str, nil, nil) {
		return false
	}

	// create index
	index_type, o := GetMysqlIndexTypeByString(field.IndexType)
	if !o {
		log.Printf("No supported index type %v", field.IndexType)
		return false
	}

	if index_type != MYSQL_INDEX_TYPE_NONE {
		index_type_length, o := GetMysqlFieldTypeDefaultLength(field.Type)
		if !o {
			log.Printf("field type %v default length not found", field.Type)
			return false
		}
		if index_type == MYSQL_INDEX_TYPE_NORMAL {

		} else if index_type == MYSQL_INDEX_TYPE_UNIQUE {

		} else if index_type == MYSQL_INDEX_TYPE_FULLTEXT {

		} else {

		}
	}

	return true
}
