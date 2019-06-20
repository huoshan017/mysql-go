package mysql_generate

import (
	"log"
	"os"
	"strings"

	"github.com/huoshan017/mysql-go/base"
)

func _upper_first_char(str string) string {
	if str == "" {
		return str
	}
	c := []byte(str)
	var uppered bool
	for i := 0; i < len(c); i++ {
		if i == 0 || c[i-1] == '_' {
			if int32(c[i]) >= int32('a') && int32(c[i]) <= int32('z') {
				c[i] = byte(int32(c[i]) + int32('A') - int32('a'))
				uppered = true
			}
		}
	}
	if !uppered {
		return str
	}
	return string(c)
}

func gen_row_func(struct_row_name string, go_type string, field *mysql_base.FieldConfig) string {
	var str string
	str += "func (this *" + struct_row_name + ") Get_" + field.Name + "() " + go_type + " {\n"
	str += "	return this." + field.Name + "\n"
	str += "}\n\n"
	str += "func (this *" + struct_row_name + ") Set_" + field.Name + "(v " + go_type + ") {\n"
	str += "	this." + field.Name + " = v\n"
	str += "}\n\n"
	str += "func (this *" + struct_row_name + ") Get_" + field.Name + "_WithLock() " + go_type + " {\n"
	str += "	this.locker.RLock()\n"
	str += "	defer this.locker.RUnlock()\n"
	str += "	return this." + field.Name + "\n"
	str += "}\n\n"
	str += "func (this *" + struct_row_name + ") Set_" + field.Name + "_WithLock(v " + go_type + ") {\n"
	str += "	this.locker.Lock()\n"
	str += "	defer this.locker.Unlock()\n"
	str += "	this." + field.Name + " = v\n"
	str += "}\n\n"
	if field.StructName != "" {
		str += "func (this *" + struct_row_name + ") Marshal_" + field.Name + "() []byte {\n"
		str += "	data, err := proto.Marshal(this." + field.Name + ")\n"
		str += "	if err != nil {\n"
		str += "		log.Printf(\"Marshal " + field.StructName + " failed err(%v)!\\n\", err.Error())\n"
		str += "		return nil\n"
		str += "	}\n"
		str += "	return data\n"
		str += "}\n\n"
		str += "func (this *" + struct_row_name + ") Unmarshal_" + field.Name + "(data []byte) bool {\n"
		str += "	err := proto.Unmarshal(data, this." + field.Name + ")\n"
		str += "	if err != nil {\n"
		str += "		log.Printf(\"Unmarshal " + field.StructName + " failed err(%v)!\\n\", err.Error())\n"
		str += "		return false\n"
		str += "	}\n"
		str += "	return true\n"
		str += "}\n\n"
		str += "func (this *" + struct_row_name + ") Get_" + field.Name + "_FVP() *mysql_base.FieldValuePair {\n"
		str += "	data := this.Marshal_" + field.Name + "()\n"
		str += "	if data == nil {\n"
		str += "		return nil\n"
		str += "	}\n"
		str += "	return &mysql_base.FieldValuePair{ Name: \"" + field.Name + "\", Value: data }\n"
		str += "}\n\n"
	} else {
		str += "func (this *" + struct_row_name + ") Get_" + field.Name + "_FVP() *mysql_base.FieldValuePair {\n"
		str += "	return &mysql_base.FieldValuePair{ Name: \"" + field.Name + "\", Value: this.Get_" + field.Name + "() }\n"
		str += "}\n\n"
	}
	return str
}

func gen_row_get_fvp_list_with_name_func(struct_row_name, field_func_map string) string {
	var str string
	str += "func (this *" + struct_row_name + ") GetFVPList(fields_name []string) []*mysql_base.FieldValuePair {\n"
	str += "	var field_list []*mysql_base.FieldValuePair\n"
	str += "	for _, field_name := range fields_name {\n"
	str += "		fun := " + field_func_map + "[field_name]\n"
	str += "		if fun == nil {\n"
	str += "			continue\n"
	str += "		}\n"
	str += "		value_pair := fun(this)\n"
	str += "		if value_pair != nil {\n"
	str += "			field_list = append(field_list, value_pair)\n"
	str += "		}\n"
	str += "	}\n"
	str += "	return field_list\n"
	str += "}\n\n"
	return str
}

func gen_row_format_all_fvp_func(struct_row_name string, table *mysql_base.TableConfig) string {
	var str string
	str += ("func (this *" + struct_row_name + ") _format_field_list() []*mysql_base.FieldValuePair {\n")
	str += ("	var field_list []*mysql_base.FieldValuePair\n")
	for _, field := range table.Fields {
		is_unsigned := strings.Contains(field.CreateFlags, "unsigned") || strings.Contains(field.CreateFlags, "UNSIGNED")
		if mysql_base.MysqlFieldTypeStr2GoTypeStr(strings.ToUpper(field.Type), is_unsigned) == "" {
			continue
		}
		if field.StructName != "" {
			str += "	data_" + field.Name + " := this.Marshal_" + field.Name + "()\n"
			str += "	if data_" + field.Name + " != nil {\n"
			str += "		field_list = append(field_list, &mysql_base.FieldValuePair{ Name: \"" + field.Name + "\", Value: data_" + field.Name + " })\n"
			str += "	}\n"
		} else {
			str += "	field_list = append(field_list, &mysql_base.FieldValuePair{ Name: \"" + field.Name + "\", Value: this.Get_" + field.Name + "() })\n"
		}
	}
	str += "	return field_list\n"
	str += "}\n\n"
	return str
}

func gen_row_lock_func(struct_row_name string) string {
	var str string
	str += "func (this *" + struct_row_name + ") Lock() {\n"
	str += "	this.locker.Lock()\n"
	str += "}\n\n"
	str += "func (this *" + struct_row_name + ") Unlock() {\n"
	str += "	this.locker.Unlock()\n"
	str += "}\n\n"
	str += "func (this *" + struct_row_name + ") RLock() {\n"
	str += "	this.locker.RLock()\n"
	str += "}\n\n"
	str += "func (this *" + struct_row_name + ") RUnlock() {\n"
	str += "	this.locker.RUnlock()\n"
	str += "}\n\n"
	var row_atomic_exec_func = struct_row_name + "_AtomicExecFunc"
	str += "type " + row_atomic_exec_func + " func(*" + struct_row_name + ")\n\n"
	str += "func (this *" + struct_row_name + ") AtomicExecute(exec_func " + row_atomic_exec_func + ") {\n"
	str += "	this.locker.Lock()\n"
	str += "	defer this.locker.Unlock()\n"
	str += "	exec_func(this)\n"
	str += "}\n\n"
	str += "func (this *" + struct_row_name + ") AtomicExecuteReadOnly(exec_func " + row_atomic_exec_func + ") {\n"
	str += "	this.locker.RLock()\n"
	str += "	defer this.locker.RUnlock()\n"
	str += "	exec_func(this)\n"
	str += "}\n\n"
	return str
}

func gen_get_result_list(table *mysql_base.TableConfig, struct_row_name, bytes_define_list, dest_list string) (str string) {
	str = ("	var r []*" + struct_row_name + "\n")
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
	return
}

func gen_get_result_map(table *mysql_base.TableConfig, struct_row_name, bytes_define_list, dest_list, primary_field_type string) (str string) {
	str = ("	var r = make(map[" + primary_field_type + "]*" + struct_row_name + ")\n")
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
	str += ("		r[t." + table.PrimaryKey + "] = t\n")
	str += ("	}\n")
	return
}

func gen_source(f *os.File, pkg_name string, table *mysql_base.TableConfig) bool {
	str := "package " + pkg_name + "\n\nimport (\n"
	if table.HasStructField() {
		str += "	\"log\"\n"
	}
	str += "	\"sync\"\n"
	str += "	\"github.com/huoshan017/mysql-go/base\"\n"
	str += "	\"github.com/huoshan017/mysql-go/manager\"\n"
	str += "	\"github.com/huoshan017/mysql-go/proxy/client\"\n"
	if table.HasStructField() {
		str += "	\"github.com/golang/protobuf/proto\"\n"
	}
	str += ")\n\n"

	struct_row_name := _upper_first_char(table.Name)
	struct_table_name := struct_row_name + "_Table"

	field_pair_func_define := table.Name + "_field_pair_func"
	field_pair_func_type := "func (t *" + struct_row_name + ") *mysql_base.FieldValuePair"
	str += "type " + field_pair_func_define + " " + field_pair_func_type + "\n\n"

	field_func_map := table.Name + "_fields_map"
	str += "var " + field_func_map + " = map[string]" + field_pair_func_define + "{\n"
	for _, field := range table.Fields {
		is_unsigned := strings.Contains(field.CreateFlags, "unsigned") || strings.Contains(field.CreateFlags, "UNSIGNED")
		if mysql_base.MysqlFieldTypeStr2GoTypeStr(strings.ToUpper(field.Type), is_unsigned) == "" {
			continue
		}
		str += "	\"" + field.Name + "\": " + field_pair_func_type + "{\n"
		str += "		return t.Get_" + field.Name + "_FVP()\n"
		str += "	},\n"
	}
	str += "}\n\n"

	// row struct
	var init_mem_list, row_func_list string
	str += ("type " + struct_row_name + " struct {\n")
	for _, field := range table.Fields {
		var go_type string
		if field.StructName != "" {
			go_type = "*" + field.StructName
			init_mem_list += "		" + field.Name + ": &" + field.StructName + "{},\n"
		} else {
			is_unsigned := strings.Contains(field.CreateFlags, "unsigned") || strings.Contains(field.CreateFlags, "UNSIGNED")
			go_type = mysql_base.MysqlFieldTypeStr2GoTypeStr(strings.ToUpper(field.Type), is_unsigned)
			if go_type == "" {
				log.Printf("get go type failed by field type %v in table %v, to continue\n", field.Type, table.Name)
				continue
			}
		}
		str += ("	" + field.Name + " " + go_type + "\n")
		row_func_list += gen_row_func(struct_row_name, go_type, field)
	}
	str += "	locker sync.RWMutex\n"
	str += "}\n\n"
	str += "func Create_" + struct_row_name + "() *" + struct_row_name + " {\n"
	str += "	return &" + struct_row_name + "{\n"
	if init_mem_list != "" {
		str += init_mem_list
	}
	str += "	}\n"
	str += "}\n\n"
	str += row_func_list
	str += gen_row_get_fvp_list_with_name_func(struct_row_name, field_func_map)
	str += gen_row_format_all_fvp_func(struct_row_name, table)
	str += gen_row_lock_func(struct_row_name)

	// table
	str += ("type " + struct_table_name + " struct {\n")
	str += "	db *mysql_manager.DB\n"
	if table.SingleRow {
		str += "	row *" + struct_row_name + "\n"
	}
	str += "}\n\n"

	// init func
	str += ("func (this *" + struct_table_name + ") Init(db *mysql_manager.DB) {\n")
	str += ("	this.db = db\n")
	str += "}\n\n"

	var field_list string
	for i, field := range table.Fields {
		is_unsigned := strings.Contains(field.CreateFlags, "unsigned") || strings.Contains(field.CreateFlags, "UNSIGNED")
		go_type := mysql_base.MysqlFieldTypeStr2GoTypeStr(strings.ToUpper(field.Type), is_unsigned)
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
		is_unsigned := strings.Contains(field.CreateFlags, "unsigned") || strings.Contains(field.CreateFlags, "UNSIGNED")
		go_type := mysql_base.MysqlFieldTypeStr2GoTypeStr(strings.ToUpper(field.Type), is_unsigned)
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
	if !table.SingleRow {
		str += ("func (this *" + struct_table_name + ") Select(field_name string, field_value interface{}) (*" + struct_row_name + ", error) {\n")
	} else {
		str += "func (this *" + struct_table_name + ") Select() (*" + struct_row_name + ", error) {\n"
	}
	str += ("	var field_list = []string{" + field_list + "}\n")
	str += ("	var t = Create_" + struct_row_name + "()\n")
	if bytes_define_list != "" {
		str += ("	var " + bytes_define_list + " []byte\n")
	}
	str += ("	var err error\n")
	str += ("	var dest_list = []interface{}{" + dest_list + "}\n")
	if !table.SingleRow {
		str += ("	err = this.db.Select(\"" + table.Name + "\", field_name, field_value, field_list, dest_list)\n")
	} else {
		str += ("	err = this.db.Select(\"" + table.Name + "\", \"place_hold\", 1, field_list, dest_list)\n")
	}
	str += ("	if err != nil {\n")
	str += ("		return nil, err\n")
	str += ("	}\n")
	for _, field := range table.Fields {
		if field.StructName != "" && (mysql_base.IsMysqlFieldBinaryType(field.RealType) || mysql_base.IsMysqlFieldBlobType(field.RealType)) {
			str += "	t.Unmarshal_" + field.Name + "(data_" + field.Name + ")\n"
		}
	}
	str += ("	return t, nil\n")
	str += ("}\n\n")

	// primary field
	var pf *mysql_base.FieldConfig
	var pt string
	if !table.SingleRow {
		pf = table.GetPrimaryKeyFieldConfig()
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
		is_unsigned := strings.Contains(pf.CreateFlags, "unsigned") || strings.Contains(pf.CreateFlags, "UNSIGNED")
		pt = mysql_base.MysqlFieldType2GoTypeStr(primary_type, is_unsigned)
		if pt == "" {
			log.Printf("主键类型%v没有对应的数据类型\n")
			return false
		}
	}

	if !table.SingleRow {
		// select records count
		str += "func (this *" + struct_table_name + ") SelectRecordsCount() (count int32, err error) {\n"
		str += "	return this.db.SelectRecordsCount(\"" + table.Name + "\")\n"
		str += "}\n\n"

		// select records count by field
		str += "func (this *" + struct_table_name + ") SelectRecordsCountByField(field_name string, field_value interface{}) (count int32, err error) {\n"
		str += "	return this.db.SelectRecordsCountByField(\"" + table.Name + "\", field_name, field_value)\n"
		str += "}\n\n"

		// select records condition
		str += "func (this *" + struct_table_name + ") SelectRecordsCondition(field_name string, field_value interface{}, sel_cond *mysql_base.SelectCondition) ([]*" + struct_row_name + ", error) {\n"
		str += "	var field_list = []string{" + field_list + "}\n"
		str += "	var result_list mysql_base.QueryResultList\n"
		str += "	err := this.db.SelectRecordsCondition(\"" + table.Name + "\", field_name, field_value, sel_cond, field_list, &result_list)\n"
		str += "	if err != nil {\n"
		str += "		return nil, err\n"
		str += "	}\n"
		str += gen_get_result_list(table, struct_row_name, bytes_define_list, dest_list)
		str += "	return r, nil\n"
		str += "}\n\n"

		// select records map condition
		str += "func (this *" + struct_table_name + ") SelectRecordsMapCondition(field_name string, field_value interface{}, sel_cond *mysql_base.SelectCondition) (map[" + pt + "]*" + struct_row_name + ", error) {\n"
		str += "	var field_list = []string{" + field_list + "}\n"
		str += "	var result_list mysql_base.QueryResultList\n"
		str += "	err := this.db.SelectRecordsCondition(\"" + table.Name + "\", field_name, field_value, sel_cond, field_list, &result_list)\n"
		str += "	if err != nil {\n"
		str += "		return nil, err\n"
		str += "	}\n"
		str += gen_get_result_map(table, struct_row_name, bytes_define_list, dest_list, pt)
		str += "	return r, nil\n"
		str += "}\n\n"

		// select all records
		str += "func (this *" + struct_table_name + ") SelectAllRecords() ([]*" + struct_row_name + ", error) {\n"
		str += "	var field_list = []string{" + field_list + "}\n"
		str += "	var result_list mysql_base.QueryResultList\n"
		str += "	err := this.db.SelectAllRecords(\"" + table.Name + "\", field_list, &result_list)\n"
		str += "	if err != nil {\n"
		str += "		return nil, err\n"
		str += "	}\n"
		str += gen_get_result_list(table, struct_row_name, bytes_define_list, dest_list)
		str += "	return r, nil\n"
		str += "}\n\n"

		// select all map records
		str += "func (this *" + struct_table_name + ") SelectAllMapRecords() (map[" + pt + "]*" + struct_row_name + ", error) {\n"
		str += "	var field_list = []string{" + field_list + "}\n"
		str += "	var result_list mysql_base.QueryResultList\n"
		str += "	err := this.db.SelectAllRecords(\"" + table.Name + "\", field_list, &result_list)\n"
		str += "	if err != nil {\n"
		str += "		return nil, err\n"
		str += "	}\n"
		str += gen_get_result_map(table, struct_row_name, bytes_define_list, dest_list, pt)
		str += "	return r, nil\n"
		str += "}\n\n"

		// select records
		str += "func (this *" + struct_table_name + ") SelectRecords(field_name string, field_value interface{}) ([]*" + struct_row_name + ", error) {\n"
		str += "	var field_list = []string{" + field_list + "}\n"
		str += "	var result_list mysql_base.QueryResultList\n"
		str += "	err := this.db.SelectRecords(\"" + table.Name + "\", field_name, field_value, field_list, &result_list)\n"
		str += "	if err != nil {\n"
		str += "		return nil, err\n"
		str += "	}\n"
		str += gen_get_result_list(table, struct_row_name, bytes_define_list, dest_list)
		str += "	return r, nil\n"
		str += "}\n\n"

		// select records map
		str += "func (this *" + struct_table_name + ") SelectRecordsMap(field_name string, field_value interface{}) (map[" + pt + "]*" + struct_row_name + ", error) {\n"
		str += "	var field_list = []string{" + field_list + "}\n"
		str += "	var result_list mysql_base.QueryResultList\n"
		str += "	err := this.db.SelectRecords(\"" + table.Name + "\", field_name, field_value, field_list, &result_list)\n"
		str += "	if err != nil {\n"
		str += "		return nil, err\n"
		str += "	}\n"
		str += gen_get_result_map(table, struct_row_name, bytes_define_list, dest_list, pt)
		str += "	return r, nil\n"
		str += "}\n\n"

		// select primary field
		str += ("func (this *" + struct_table_name + ") SelectByPrimaryField(key " + pt + ") (*" + struct_row_name + ", error) {\n")
		str += ("	v, err := this.Select(\"" + pf.Name + "\", key)\n")
		str += ("	if err != nil {\n")
		str += ("		return nil, err\n")
		str += ("	}\n")
		str += ("	return v, nil\n")
		str += ("}\n\n")

		// select all primary field
		str += ("func (this *" + struct_table_name + ") SelectAllPrimaryField() ([]" + pt + ", error) {\n")
		str += ("	var result_list mysql_base.QueryResultList\n")
		str += ("	err := this.db.SelectFieldNoKey(\"" + table.Name + "\", \"" + pf.Name + "\", &result_list)\n")
		str += ("	if err != nil {\n")
		str += ("		return nil, err\n")
		str += ("	}\n")
		str += ("	var value_list []" + pt + "\n")
		str += ("	for {\n")
		str += ("		var d " + pt + "\n")
		str += ("		if !result_list.Get(&d) {\n")
		str += ("			break\n")
		str += ("		}\n")
		str += ("		value_list = append(value_list, d)\n")
		str += ("	}\n")
		str += ("	return value_list, nil\n")
		str += ("}\n\n")

		// select all primary field map
		str += ("func (this *" + struct_table_name + ") SelectAllPrimaryFieldMap() (map[" + pt + "]bool, error) {\n")
		str += ("	var result_list mysql_base.QueryResultList\n")
		str += ("	err := this.db.SelectFieldNoKey(\"" + table.Name + "\", \"" + pf.Name + "\", &result_list)\n")
		str += ("	if err != nil {\n")
		str += ("		return nil, err\n")
		str += ("	}\n")
		str += ("	var value_map = make(map[" + pt + "]bool)\n")
		str += ("	for {\n")
		str += ("		var d " + pt + "\n")
		str += ("		if !result_list.Get(&d) {\n")
		str += ("			break\n")
		str += ("		}\n")
		str += ("		value_map[d] = true\n")
		str += ("	}\n")
		str += ("	return value_map, nil\n")
		str += ("}\n\n")

		// insert
		str += "func (this *" + struct_table_name + ") Insert(t *" + struct_row_name + ") {\n"
		str += "	var field_list = t._format_field_list()\n"
		str += "	if field_list != nil {\n"
		str += "		this.db.Insert(\"" + table.Name + "\", field_list)\n"
		str += "	}\n"
		str += "}\n\n"

		// insert ignore
		str += "func (this *" + struct_table_name + ") InsertIgnore(t *" + struct_row_name + ") {\n"
		str += "	var field_list = t._format_field_list()\n"
		str += "	if field_list != nil {\n"
		str += "		this.db.InsertIgnore(\"" + table.Name + "\", field_list)\n"
		str += "	}\n"
		str += "}\n\n"

		// delete
		str += ("func (this *" + struct_table_name + ") Delete(" + pf.Name + " " + pt + ") {\n")
		str += ("	this.db.Delete(\"" + table.Name + "\", \"" + pf.Name + "\", " + pf.Name + ")\n")
		str += ("}\n\n")

		// create row func
		str += "func (this *" + struct_table_name + ") NewRecord(" + pf.Name + " " + pt + ") *" + struct_row_name + " {\n"
		str += "	return &" + struct_row_name + "{ " + pf.Name + ": " + pf.Name + ", }\n"
		str += "}\n\n"
	} else {
		str += "func (this *" + struct_table_name + ") GetRow() (*" + struct_row_name + ", error) {\n"
		str += "	if this.row == nil {\n"
		str += "		row, err := this.Select()\n"
		str += "		if err != nil {\n"
		str += "			return nil, err\n"
		str += "		}\n"
		str += "		this.row = row\n"
		str += "	}\n"
		str += "	return this.row, nil\n"
		str += "}\n\n"
	}

	// update
	str += "func (this *" + struct_table_name + ") UpdateAll(t *" + struct_row_name + ") {\n"
	str += "	var field_list = t._format_field_list()\n"
	str += "	if field_list != nil {\n"
	if !table.SingleRow {
		str += "		this.db.Update(\"" + table.Name + "\", \"" + pf.Name + "\", t.Get_" + pf.Name + "(), field_list)\n"
	} else {
		str += "		this.db.Update(\"" + table.Name + "\", \"place_hold\", 1, field_list)\n"
	}
	str += "	}\n"
	str += "}\n\n"

	// update some field
	if !table.SingleRow {
		str += "func (this *" + struct_table_name + ") UpdateWithFVPList(" + pf.Name + " " + pt + ", field_list []*mysql_base.FieldValuePair) {\n"
		str += "	this.db.Update(\"" + table.Name + "\", \"" + pf.Name + "\", " + pf.Name + ", field_list)\n"
	} else {
		str += "func (this *" + struct_table_name + ") UpdateWithFVPList(field_list []*mysql_base.FieldValuePair) {\n"
		str += "	this.db.Update(\"" + table.Name + "\", \"place_hold\", 1, field_list)\n"
	}
	str += "}\n\n"

	// update by field name
	str += "func (this *" + struct_table_name + ") UpdateWithFieldName(t *" + struct_row_name + ", fields_name []string) {\n"
	str += "	var field_list = t.GetFVPList(fields_name)\n"
	str += "	if field_list != nil {\n"
	if !table.SingleRow {
		str += "		this.UpdateWithFVPList(t.Get_" + pf.Name + "(), field_list)\n"
	} else {
		str += "		this.UpdateWithFVPList(field_list)\n"
	}
	str += "	}\n"
	str += "}\n\n"

	str += gen_procedure_source(table, struct_table_name, struct_row_name, pf, pt)

	_, err := f.WriteString(str)
	if err != nil {
		log.Printf("write string err %v\n", err.Error())
		return false
	}

	return true
}

func gen_procedure_source(table *mysql_base.TableConfig, struct_table_name, struct_row_name string, primary_field *mysql_base.FieldConfig, primary_type string) string {
	var str string

	if !table.SingleRow {
		str += "func (this *" + struct_table_name + ") TransactionInsert(transaction *mysql_manager.Transaction, t *" + struct_row_name + ") {\n"
		str += "	field_list := t._format_field_list()\n"
		str += "	if field_list != nil {\n"
		str += "		transaction.Insert(\"" + table.Name + "\", field_list)\n"
		str += "	}\n"
		str += "}\n\n"
		str += "func (this *" + struct_table_name + ") TransactionDelete(transaction *mysql_manager.Transaction, " + primary_field.Name + " " + primary_type + ") {\n"
		str += "	transaction.Delete(\"" + table.Name + "\", \"" + primary_field.Name + "\", " + primary_field.Name + ")\n"
		str += "}\n\n"
	}

	str += "func (this *" + struct_table_name + ") TransactionUpdateAll(transaction *mysql_manager.Transaction, t*" + struct_row_name + ") {\n"
	str += "	field_list := t._format_field_list()\n"
	str += "	if field_list != nil {\n"
	if !table.SingleRow {
		str += "		transaction.Update(\"" + table.Name + "\", \"" + primary_field.Name + "\", t.Get_" + primary_field.Name + "(), field_list)\n"
	} else {
		str += "		transaction.Update(\"" + table.Name + "\", \"place_hold\", 1, field_list)\n"
	}
	str += "	}\n"
	str += "}\n\n"

	if !table.SingleRow {
		str += "func (this *" + struct_table_name + ") TransactionUpdateWithFVPList(transaction *mysql_manager.Transaction, " + primary_field.Name + " " + primary_type + ", field_list []*mysql_base.FieldValuePair) {\n"
		str += "	transaction.Update(\"" + table.Name + "\", \"" + primary_field.Name + "\", " + primary_field.Name + ", field_list)\n"
	} else {
		str += "func (this *" + struct_table_name + ") TransactionUpdateWithFVPList(transaction *mysql_manager.Transaction, field_list []*mysql_base.FieldValuePair) {\n"
		str += "	transaction.Update(\"" + table.Name + "\", \"place_hold\", 1, field_list)\n"
	}
	str += "}\n\n"

	str += "func (this *" + struct_table_name + ") TransactionUpdateWithFieldName(transaction *mysql_manager.Transaction, t *" + struct_row_name + ", fields_name []string) {\n"
	str += "	field_list := t.GetFVPList(fields_name)\n"
	str += "	if field_list != nil {\n"
	if !table.SingleRow {
		str += "		transaction.Update(\"" + table.Name + "\", \"" + primary_field.Name + "\", t.Get_" + primary_field.Name + "(), field_list)\n"
	} else {
		str += "		transaction.Update(\"" + table.Name + "\", \"place_hold\", 1, field_list)\n"
	}
	str += "	}\n"
	str += "}\n\n"

	return str
}
