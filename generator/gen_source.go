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
	if table.HasStructField() {
		str += "	\"log\"\n"
	}
	str += "	\"github.com/huoshan017/mysql-go/base\"\n"
	str += "	\"github.com/huoshan017/mysql-go/manager\"\n"
	if table.HasStructField() {
		str += "	\"github.com/golang/protobuf/proto\"\n"
	}
	str += ")\n\n"

	var init_mem_list string
	var row_func_list string
	// row struct
	struct_row_name := _upper_first_char(table.Name)
	str += ("type " + struct_row_name + " struct {\n")
	for _, field := range table.Fields {
		var go_type string
		if field.StructName != "" {
			go_type = "*" + field.StructName
			init_mem_list += "		" + field.Name + ": &" + field.StructName + "{},\n"
		} else {
			go_type = _field_type_string_to_go_type(strings.ToUpper(field.Type))
			if go_type == "" {
				log.Printf("get go type failed by field type %v in table %v\n", field.Type, table.Name)
				// 跳过不处理
				continue
			}
		}
		str += ("	" + field.Name + " " + go_type + "\n")
		row_func_list += ("func (this *" + struct_row_name + ") Get_" + field.Name + "() " + go_type + " {\n")
		row_func_list += ("	return this." + field.Name + "\n")
		row_func_list += ("}\n\n")
		row_func_list += ("func (this *" + struct_row_name + ") Set_" + field.Name + "(v " + go_type + ") {\n")
		row_func_list += ("	this." + field.Name + " = v\n")
		row_func_list += ("}\n\n")
	}
	str += "}\n\n"
	str += "func Create_" + struct_row_name + "() *" + struct_row_name + " {\n"
	str += "	return &" + struct_row_name + "{\n"
	if init_mem_list != "" {
		str += init_mem_list
	}
	str += "	}\n"
	str += "}\n\n"
	str += row_func_list

	// table
	struct_table_name := struct_row_name + "Table"
	str += ("type " + struct_table_name + " struct {\n")
	str += "	db *mysql_manager.DB\n"
	pf := table.GetPrimaryKeyFieldConfig()
	if pf == nil {
		log.Printf("cant get table %v primary key\n", table.Name)
		return false
	}
	primary_type, o := mysql_base.GetMysqlFieldTypeByString(strings.ToUpper(pf.Type))
	if !o {
		log.Printf("table %v primary type %v invalid", table.Name, pf.Type)
		return false
	}
	if !(mysql_base.IsMysqlFieldIntType(primary_type) || mysql_base.IsMysqlFieldTextType(primary_type)) {
		log.Printf("not support primary type %v for table %v", pf.Type, table.Name)
		return false
	}
	pt := _field_type_to_go_type(primary_type)
	if pt == "" {
		log.Printf("主键类型%v没有对应的数据类型\n")
		return false
	}
	str += "	rows map[" + pt + "]*" + struct_row_name + "\n"
	str += "}\n\n"

	// init func
	str += ("func (this *" + struct_table_name + ") Init(db *mysql_manager.DB) {\n")
	str += ("	this.db = db\n")
	str += "}\n\n"

	var field_list string
	for i, field := range table.Fields {
		go_type := _field_type_string_to_go_type(strings.ToUpper(field.Type))
		if go_type == "" {
			continue
		}
		if i == 0 {
			field_list = "\"" + field.Name + "\""
		} else {
			field_list += (", \"" + field.Name + "\"")
		}
	}

	var bytes_define_list string
	var dest_list string
	var unmarshal_bytes_list string
	for i, field := range table.Fields {
		go_type := _field_type_string_to_go_type(strings.ToUpper(field.Type))
		if go_type == "" {
			continue
		}

		var dest string
		if field.StructName != "" && (mysql_base.IsMysqlFieldBinaryType(field.RealType) || mysql_base.IsMysqlFieldBlobType(field.RealType)) {
			dest = "data_" + field.Name
			if bytes_define_list == "" {
				bytes_define_list = dest
			} else {
				bytes_define_list += (", " + dest)
			}
			if unmarshal_bytes_list == "" {
				unmarshal_bytes_list += "	var err error\n"
			}
			unmarshal_bytes_list += "	err = proto.Unmarshal(" + dest + ", v." + field.Name + ")\n"
			unmarshal_bytes_list += "	if err != nil {\n"
			unmarshal_bytes_list += "		log.Printf(\"Unmarshal " + field.StructName + " failed err(%s)!\\n\", err.Error())\n"
			unmarshal_bytes_list += "	}\n"
		} else {
			dest = "v." + field.Name
		}

		if i == 0 {
			if mysql_base.IsMysqlFieldBinaryType(field.RealType) || mysql_base.IsMysqlFieldBlobType(field.RealType) {
				dest_list = "&" + dest
			} else {
				dest_list = "&" + dest
			}
		} else {
			if mysql_base.IsMysqlFieldBinaryType(field.RealType) || mysql_base.IsMysqlFieldBlobType(field.RealType) {
				dest_list += (", &" + dest)
			} else {
				dest_list += (", &" + dest)
			}
		}
	}

	// select func
	str += ("func (this *" + struct_table_name + ") Select(key string, value interface{}) (*" + struct_row_name + ", bool) {\n")
	str += ("	var field_list = []string{" + field_list + "}\n")
	str += ("	var v = Create_" + struct_row_name + "()\n")
	if bytes_define_list != "" {
		str += ("	var " + bytes_define_list + " []byte\n")
	}
	str += ("	var dest_list = []interface{}{" + dest_list + "}\n")
	str += ("	if !this.db.Select(\"" + table.Name + "\", key, value, field_list, dest_list) {\n")
	str += ("		return nil, false\n")
	str += ("	}\n")
	if unmarshal_bytes_list != "" {
		str += unmarshal_bytes_list
	}
	str += ("	return v, true\n")
	str += ("}\n\n")

	// select multi func
	str += ("func (this *" + struct_table_name + ") SelectMulti(key string, value interface{}) ([]*" + struct_row_name + ", bool) {\n")
	str += ("	var field_list = []string{" + field_list + "}\n")
	str += ("	var result_list mysql_base.QueryResultList\n")
	str += ("	if !this.db.SelectRecords(\"" + table.Name + "\", key, value, field_list, &result_list) {\n")
	str += ("		return nil, false\n")
	str += ("	}\n")
	str += ("	var r []*" + struct_row_name + "\n")
	if bytes_define_list != "" {
		str += ("	var " + bytes_define_list + " []byte\n")
	}
	str += ("	for {\n")
	str += ("		var v = Create_" + struct_row_name + "()\n")
	str += ("		var dest_list = []interface{}{" + dest_list + "}\n")
	str += ("		if !result_list.Get(dest_list...) {\n")
	str += ("			break\n")
	str += ("		}\n")
	if unmarshal_bytes_list != "" {
		str += unmarshal_bytes_list
	}
	str += ("		r = append(r, v)\n")
	str += ("	}\n")
	str += ("	return r, true\n")
	str += ("}\n\n")

	_, err := f.WriteString(str)
	if err != nil {
		log.Printf("write string err %v\n", err.Error())
		return false
	}

	return true
}
