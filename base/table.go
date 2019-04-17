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
	sql_str := fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%s` (`%s` %s(%v) %s, PRIMARY KEY(`%s`)) ENGINE=%s", tab.Name, tab.PrimaryKey, primary_field.Type, primary_field.Length /*primary_field.CreateFlags*/, strings.Replace(primary_field.CreateFlags, ",", " ", -1), tab.PrimaryKey, tab.Engine)
	if !this.Exec(sql_str, nil, nil) {
		return false
	}

	// add fields
	for _, f := range tab.Fields {
		if tab.PrimaryKey == f.Name {
			continue
		}
		if !this.add_field(tab.Name, f) {
			return false
		}
	}

	return true
}

func (this *Database) DropTable(table_name string) bool {
	args := []interface{}{table_name}
	if !this.ExecWith("DROP TABLE ?", args, nil, nil) {
		return false
	}
	return true
}

func _get_field_type_str(field *FieldConfig) (field_type_str string) {
	field_type_str = strings.ToUpper(field.Type)
	field_type, o := GetMysqlFieldTypeByString(field_type_str)
	if !o {
		log.Printf("Cant get field %v type string", field.Name)
		return ""
	}

	if field_type == MYSQL_FIELD_TYPE_DATE || field_type == MYSQL_FIELD_TYPE_DATETIME || (field_type == MYSQL_FIELD_TYPE_TIMESTAMP && field.Length == MYSQL_FIELD_DEFAULT_LENGTH) {
		return
	}

	if field.Length == MYSQL_FIELD_DEFAULT_LENGTH {
		default_length, o := GetMysqlFieldTypeDefaultLength(field_type)
		if !o {
			log.Printf("Cant get field type %v default length", field_type)
			return ""
		}
		field_type_str = fmt.Sprintf("%v(%v)", field_type_str, default_length)
	} else {
		field_type_str = fmt.Sprintf("%v(%v)", field_type_str, field.Length)
	}

	return
}

func _get_field_create_flags_str(field *FieldConfig) (create_flags_str string) {
	flags := strings.Split(field.CreateFlags, ",")
	for _, s := range flags {
		if s == "" {
			continue
		}
		s = strings.ToUpper(s)
		t, o := GetMysqlTableCreateFlagTypeByString(s)
		if !o {
			log.Printf("Create table flag %v not found\n", s)
			break
		}
		// 缺省
		if t == MYSQL_TABLE_CREATE_DEFAULT {
			field_type, _ := GetMysqlFieldTypeByString(strings.ToUpper(field.Type))
			if IsMysqlFieldIntType(field_type) {
				create_flags_str += (" " + s + " 0")
			} else if IsMysqlFieldTextType(field_type) || IsMysqlFieldBinaryType(field_type) || IsMysqlFieldBlobType(field_type) {
				create_flags_str += (" " + s + " ''")
			} else if IsMysqlFieldTimestampType(field_type) {
				create_flags_str += (" " + s + " CURRENT_TIMESTAMP")
			} else {
				log.Printf("Create table flag %v no default value", s)
			}
		} else {
			create_flags_str += (" " + s)
		}
	}
	return
}

func (this *Database) add_field(table_name string, field *FieldConfig) bool {
	sql_str := fmt.Sprintf("DESCRIBE %s %s", table_name, field.Name)
	if this.HasRow(sql_str) {
		log.Printf("describe get rows not empty, no need to add field %v\n", field.Name)
		return true
	}

	field_type_str := _get_field_type_str(field)
	if field_type_str == "" {
		return false
	}

	create_flags_str := _get_field_create_flags_str(field)

	sql_str = fmt.Sprintf("ALTER TABLE `%s` ADD COLUMN `%s` %s %s", table_name, field.Name, field_type_str, create_flags_str)
	if !this.Exec(sql_str, nil, nil) {
		log.Printf("create table %v field %v failed\n", table_name, field.Name)
		return false
	}

	// create index
	index_type, o := GetMysqlIndexTypeByString(strings.ToUpper(field.IndexType))
	if !o {
		log.Printf("No supported index type %v\n", field.IndexType)
		return false
	}

	if index_type != MYSQL_INDEX_TYPE_NONE {
		if index_type == MYSQL_INDEX_TYPE_NORMAL {
			sql_str = fmt.Sprintf("ALTER TABLE `%s` ADD INDEX %s_index(`%s`)", table_name, field.Name, field.Name)
		} else if index_type == MYSQL_INDEX_TYPE_UNIQUE {
			sql_str = fmt.Sprintf("ALTER TABLE `%s` ADD UNIQUE (`%s`)", table_name, field.Name)
		} else if index_type == MYSQL_INDEX_TYPE_FULLTEXT {
			sql_str = fmt.Sprintf("ALTER TABLE `%s` ADD FULLTEXT(`%s`)", table_name, field.Name)
		} else {
			log.Printf("table %v field %v index type FULLTEXT not supported\n", table_name, field.Name)
		}

		if !this.Exec(sql_str, nil, nil) {
			log.Printf("create table %v field %v index failed\n", table_name, field.Name)
			return false
		}
	}

	return true
}

func (this *Database) remove_field(table_name, field_name string) bool {
	args := []interface{}{table_name, field_name}
	if !this.ExecWith("ALTER TABLE ? DROP COLUMN ?", args, nil, nil) {
		return false
	}
	return true
}

func (this *Database) rename_field(table_name, old_field_name, new_field_name string) bool {
	args := []interface{}{table_name, old_field_name, new_field_name}
	if !this.ExecWith("ALTER TABLE ? CHANGE ? ?", args, nil, nil) {
		return false
	}
	return true
}

func (this *Database) modify_field_attr(table_name string, field *FieldConfig) bool {
	field_type_str := _get_field_type_str(field)
	if field_type_str == "" {
		log.Printf("get table %v field %v type string failed\n", table_name, field.Name)
		return false
	}

	field_create_str := _get_field_create_flags_str(field)
	args := []interface{}{table_name, field.Name, field_type_str, field_create_str}
	if !this.ExecWith("ALTER TABLE ? MODIFY ? ? ?", args, nil, nil) {
		log.Printf("modify table %v field %v attr failed\n", table_name, field.Name)
		return false
	}
	return true
}
