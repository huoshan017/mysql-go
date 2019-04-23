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
		if field.StructName != "" {
			row_func_list += "func (this *" + struct_row_name + ") Marshal_" + field.Name + "() []byte {\n"
			row_func_list += "	data, err := proto.Marshal(this." + field.Name + ")\n"
			row_func_list += "	if err != nil {\n"
			row_func_list += "		log.Printf(\"Marshal " + field.StructName + " failed err(%v)!\\n\", err.Error())\n"
			row_func_list += "		return nil\n"
			row_func_list += "	}\n"
			row_func_list += "	return data\n"
			row_func_list += "}\n\n"
			row_func_list += "func (this *" + struct_row_name + ") Unmarshal_" + field.Name + "(data []byte) bool {\n"
			row_func_list += "	err := proto.Unmarshal(data, this." + field.Name + ")\n"
			row_func_list += "	if err != nil {\n"
			row_func_list += "		log.Printf(\"Unmarshal " + field.StructName + " failed err(%v)!\\n\", err.Error())\n"
			row_func_list += "		return false\n"
			row_func_list += "	}\n"
			row_func_list += "	return true\n"
			row_func_list += "}\n\n"
			row_func_list += "func (this *" + struct_row_name + ") GetValuePair_" + field.Name + "() *mysql_base.FieldValuePair {\n"
			row_func_list += "	data := this.Marshal_" + field.Name + "()\n"
			row_func_list += "	if data == nil {\n"
			row_func_list += "		return nil\n"
			row_func_list += "	}\n"
			row_func_list += "	return &mysql_base.FieldValuePair{ Name: \"" + field.Name + "\", Value: data }\n"
			row_func_list += "}\n\n"
		} else {
			row_func_list += "func (this *" + struct_row_name + ") GetValuePair_" + field.Name + "() *mysql_base.FieldValuePair {\n"
			row_func_list += "	return &mysql_base.FieldValuePair{ Name: \"" + field.Name + "\", Value: this.Get_" + field.Name + "() }\n"
			row_func_list += "}\n\n"
		}
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
	struct_table_name := struct_row_name + "_Table"
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
	for _, field := range table.Fields {
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
		} else {
			dest = "t." + field.Name
		}

		if dest_list == "" {
			dest_list = "&" + dest
		} else {
			dest_list += (", &" + dest)
		}
	}

	// select func
	str += ("func (this *" + struct_table_name + ") Select(key string, value interface{}) (*" + struct_row_name + ", bool) {\n")
	str += ("	var field_list = []string{" + field_list + "}\n")
	str += ("	var t = Create_" + struct_row_name + "()\n")
	if bytes_define_list != "" {
		str += ("	var " + bytes_define_list + " []byte\n")
	}
	str += ("	var dest_list = []interface{}{" + dest_list + "}\n")
	str += ("	if !this.db.Select(\"" + table.Name + "\", key, value, field_list, dest_list) {\n")
	str += ("		return nil, false\n")
	str += ("	}\n")
	for _, field := range table.Fields {
		if field.StructName != "" && (mysql_base.IsMysqlFieldBinaryType(field.RealType) || mysql_base.IsMysqlFieldBlobType(field.RealType)) {
			str += "	t.Unmarshal_" + field.Name + "(data_" + field.Name + ")\n"
		}
	}
	str += ("	return t, true\n")
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
	str += ("		var t = Create_" + struct_row_name + "()\n")
	str += ("		var dest_list = []interface{}{" + dest_list + "}\n")
	str += ("		if !result_list.Get(dest_list...) {\n")
	str += ("			break\n")
	str += ("		}\n")
	for _, field := range table.Fields {
		if field.StructName != "" && (mysql_base.IsMysqlFieldBinaryType(field.RealType) || mysql_base.IsMysqlFieldBlobType(field.RealType)) {
			str += "		t.Unmarshal_" + field.Name + "(data_" + field.Name + ")\n"
		}
	}
	str += ("		r = append(r, t)\n")
	str += ("	}\n")
	str += ("	return r, true\n")
	str += ("}\n\n")

	// select primary field
	str += ("func (this *" + struct_table_name + ") SelectPrimaryField() ([]" + pt + ") {\n")
	str += ("	var result_list mysql_base.QueryResultList\n")
	str += ("	if !this.db.SelectFieldNoKey(\"" + table.Name + "\", \"" + pf.Name + "\", &result_list) {\n")
	str += ("		return nil\n")
	str += ("	}\n")
	str += ("	var value_list []" + pt + "\n")
	str += ("	for {\n")
	str += ("		var d " + pt + "\n")
	str += ("		if !result_list.Get(&d) {\n")
	str += ("			break\n")
	str += ("		}\n")
	str += ("		value_list = append(value_list, d)\n")
	str += ("	}\n")
	str += ("	return value_list\n")
	str += ("}\n\n")

	// _format_field_list
	str += ("func (this *" + struct_table_name + ") _format_field_list(t * " + struct_row_name + ") []*mysql_base.FieldValuePair {\n")
	str += ("	var field_list []*mysql_base.FieldValuePair\n")
	for _, field := range table.Fields {
		if _field_type_string_to_go_type(strings.ToUpper(field.Type)) == "" {
			continue
		}
		if field.StructName != "" {
			str += "	data_" + field.Name + " := t.Marshal_" + field.Name + "()\n"
			str += "	if data_" + field.Name + " != nil {\n"
			str += "		field_list = append(field_list, &mysql_base.FieldValuePair{ Name: \"" + field.Name + "\", Value: data_" + field.Name + " })\n"
			str += "	}\n"
		} else {
			str += "	field_list = append(field_list, &mysql_base.FieldValuePair{ Name: \"" + field.Name + "\", Value: t.Get_" + field.Name + "() })\n"
		}
	}
	str += "	return field_list\n"
	str += "}\n\n"

	// insert
	str += ("func (this *" + struct_table_name + ") Insert(t *" + struct_row_name + ") {\n")
	str += ("	var field_list = this._format_field_list(t)\n")
	str += ("	this.db.Insert(\"" + table.Name + "\", field_list)\n")
	str += ("}\n\n")

	// delete
	str += ("func (this *" + struct_table_name + ") Delete(" + pf.Name + " " + pt + ") {\n")
	str += ("	this.db.Delete(\"" + table.Name + "\", \"" + pf.Name + "\", " + pf.Name + ")\n")
	str += ("}\n\n")

	// update
	str += "func (this *" + struct_table_name + ") UpdateAll(" + pf.Name + " " + pt + ", t *" + struct_row_name + ") {\n"
	str += "	var field_list = this._format_field_list(t)\n"
	str += "	this.db.Update(\"" + table.Name + "\", \"" + pf.Name + "\", " + pf.Name + ", field_list)\n"
	str += "}\n\n"

	// update some field
	str += "func (this *" + struct_table_name + ") UpdateSome(" + pf.Name + " " + pt + ", field_list []*mysql_base.FieldValuePair) {\n"
	str += "	this.db.Update(\"" + table.Name + "\", \"" + pf.Name + "\", " + pf.Name + ", field_list)\n"
	str += "}\n"

	_, err := f.WriteString(str)
	if err != nil {
		log.Printf("write string err %v\n", err.Error())
		return false
	}

	return true
}
