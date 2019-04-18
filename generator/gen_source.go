package mysql_generator

import (
	"log"
	"os"
	"strings"

	"github.com/huoshan017/mysql-go/base"
)

var field_type_string_maps = map[int]string{
	mysql_base.MYSQL_FIELD_TYPE_TINYINT:    "int8",
	mysql_base.MYSQL_FIELD_TYPE_SMALLINT:   "int16",
	mysql_base.MYSQL_FIELD_TYPE_MEDIUMINT:  "int16",
	mysql_base.MYSQL_FIELD_TYPE_INT:        "int32",
	mysql_base.MYSQL_FIELD_TYPE_BIGINT:     "int64",
	mysql_base.MYSQL_FIELD_TYPE_FLOAT:      "float32",
	mysql_base.MYSQL_FIELD_TYPE_DOUBLE:     "float64",
	mysql_base.MYSQL_FIELD_TYPE_DATE:       "",
	mysql_base.MYSQL_FIELD_TYPE_DATETIME:   "",
	mysql_base.MYSQL_FIELD_TYPE_TIMESTAMP:  "",
	mysql_base.MYSQL_FIELD_TYPE_TIME:       "",
	mysql_base.MYSQL_FIELD_TYPE_YEAR:       "",
	mysql_base.MYSQL_FIELD_TYPE_CHAR:       "string",
	mysql_base.MYSQL_FIELD_TYPE_VARCHAR:    "string",
	mysql_base.MYSQL_FIELD_TYPE_TINYTEXT:   "string",
	mysql_base.MYSQL_FIELD_TYPE_MEDIUMTEXT: "string",
	mysql_base.MYSQL_FIELD_TYPE_TEXT:       "string",
	mysql_base.MYSQL_FIELD_TYPE_LONGTEXT:   "string",
	mysql_base.MYSQL_FIELD_TYPE_BINARY:     "[]byte",
	mysql_base.MYSQL_FIELD_TYPE_VARBINARY:  "[]byte",
	mysql_base.MYSQL_FIELD_TYPE_TINYBLOB:   "[]byte",
	mysql_base.MYSQL_FIELD_TYPE_MEDIUMBLOB: "[]byte",
	mysql_base.MYSQL_FIELD_TYPE_BLOB:       "[]byte",
	mysql_base.MYSQL_FIELD_TYPE_LONGBLOB:   "[]byte",
	mysql_base.MYSQL_FIELD_TYPE_ENUM:       "",
	mysql_base.MYSQL_FIELD_TYPE_SET:        "",
}

func _field_type_to_go_type(field_type int) string {
	go_type, o := field_type_string_maps[field_type]
	if !o {
		go_type = ""
	}
	return go_type
}

func _field_type_string_to_go_type(field_type_str string) string {
	field_type, o := mysql_base.GetMysqlFieldTypeByString(field_type_str)
	if !o {
		return ""
	}
	return _field_type_to_go_type(field_type)
}

func _upper_first_char(str string) string {
	if str == "" {
		return str
	}
	c := []byte(str)
	var uppered bool
	if int32(c[0]) >= int32('a') && int32(c[0]) <= int32('z') {
		c[0] = byte(int32(c[0]) + int32('A') - int32('a'))
		uppered = true
	}
	if !uppered {
		return str
	}
	return string(c)
}

func gen_source(f *os.File, pkg_name string, table *mysql_base.TableConfig) bool {
	str := "package " + pkg_name + "\n\nimport (\n"
	str += "	\"io/ioutil\"\n"
	str += "	\"log\"\n"
	str += "	\"strconv\"\n"
	str += "	\"strings\"\n"
	str += "	\"github.com/huoshan017/mysql-go/base\"\n"
	str += ")\n\n"

	// row struct
	row_name := _upper_first_char(table.Name)
	struct_row_name := row_name + "Row"
	str += ("type " + struct_row_name + " struct {\n")
	for _, field := range table.Fields {
		field_type, o := mysql_base.GetMysqlFieldTypeByString(strings.ToUpper(field.Type))
		if !o {
			log.Printf("cant get field type by string %v\n", field.Type)
			return false
		}
		go_type := _field_type_to_go_type(field_type)
		if go_type == "" {
			log.Printf("get go type failed by field type %v", field_type)
			return false
		}
		str += ("	" + field.Name + " " + go_type + "\n")
	}
	str += "}\n\n"

	// table
	struct_table_name := row_name + "Table"
	str += ("type " + struct_table_name + " struct {\n")
	str += "	db *mysql_base.Database\n"
	pf := table.GetPrimaryKeyFieldConfig()
	if pf == nil {
		log.Printf("cant get table %v primary key\n", table.Name)
		return false
	}
	primary_type, o := mysql_base.GetMysqlFieldTypeByString(pf.Type)
	if !o {
		log.Printf("table %v primary type invalid", table.Name, pf.Type)
		return false
	}
	if !(mysql_base.IsMysqlFieldIntType(primary_type) || mysql_base.IsMysqlFieldTextType(primary_type)) {
		log.Printf("not support primary type %v for table %v", pf.Type, table.Name)
		return false
	}
	pt := _field_type_to_go_type(primary_type)
	str += "	rows map[" + pt + "]*" + struct_row_name + "\n"
	str += "}\n\n"

	// init func
	str += ("func (this *" + struct_table_name + ") Init(db *mysql_base.Database) {\n")
	str += ("	this.db = db\n")
	str += "}\n\n"

	// select func
	str += ("func (this *" + struct_table_name + ") Select()")

	_, err := f.WriteString(str)
	if err != nil {
		log.Printf("write string err %v\n", err.Error())
		return false
	}

	return true
}
