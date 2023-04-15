package mysql_generate

import (
	"os"
	"strconv"
	"strings"

	mysql_base "github.com/huoshan017/mysql-go/base"
	"github.com/huoshan017/mysql-go/log"
)

func gen_get_proxy_result_list(table *mysql_base.TableConfig, struct_row_name, bytes_define_list, dest_list string) (str string) {
	str = ("	var r []*" + struct_row_name + "\n")
	if bytes_define_list != "" {
		str += ("	var " + bytes_define_list + " []byte\n")
	}
	str += ("	for {\n")
	str += ("		var t = Create" + struct_row_name + "()\n")
	str += ("		var dest_list = []interface{}{" + dest_list + "}\n")
	str += ("		if !result_list.Get(dest_list...) {\n")
	str += ("			break\n")
	str += ("		}\n")
	for _, field := range table.Fields {
		if field.StructName != "" && (mysql_base.IsMysqlFieldBinaryType(field.Type) || mysql_base.IsMysqlFieldBlobType(field.Type)) {
			str += "		t.Unmarshal_" + field.Name + "(data_" + field.Name + ")\n"
		}
	}
	str += ("		r = append(r, t)\n")
	str += ("	}\n")
	return
}

func gen_get_proxy_result_map(table *mysql_base.TableConfig, struct_row_name, bytes_define_list, dest_list, primary_field_type string) (str string) {
	str = ("	var r = make(map[" + primary_field_type + "]*" + struct_row_name + ")\n")
	if bytes_define_list != "" {
		str += ("	var " + bytes_define_list + " []byte\n")
	}
	str += ("	for k, v := range records_map {\n")
	str += ("		var t = Create" + struct_row_name + "()\n")
	for i, field := range table.Fields {
		if field.StructName != "" && (mysql_base.IsMysqlFieldBinaryType(field.Type) || mysql_base.IsMysqlFieldBlobType(field.Type)) {
			str += "		t.Unmarshal_" + field.Name + "(data_" + field.Name + ")\n"
		} else if mysql_base.IsMysqlFieldIntType(field.Type) || mysql_base.IsMysqlFieldTextType(field.Type) {
			str += "		mysql_base.CopySrcValue2Dest(&t." + field.Name + ", v[" + strconv.Itoa(i) + "])\n"
		}
	}
	str += ("		r[k.(" + primary_field_type + ")] = t\n")
	str += ("	}\n")
	return
}

func get_primary_field_and_type(table *mysql_base.TableConfig) (*mysql_base.FieldConfig, string, bool) {
	pf := table.GetPrimaryKeyFieldConfig()
	if pf == nil {
		log.Infof("cant get table %v primary key", table.Name)
		return nil, "", false
	}
	if !(mysql_base.IsMysqlFieldIntType(pf.Type) || mysql_base.IsMysqlFieldTextType(pf.Type)) {
		log.Infof("not support primary type %v for table %v", pf.Type, table.Name)
		return nil, "", false
	}
	isUnsigned := strings.Contains(strings.ToLower(pf.TypeStr), "unsigned")
	pt := mysql_base.MysqlFieldType2GoTypeStr(pf.Type, isUnsigned)
	if pt == "" {
		log.Infof("主键类型%v没有对应的数据类型", pt)
		return nil, "", false
	}
	return pf, pt, true
}

func gen_proxy_source(f *os.File, pkg_name string, table *mysql_base.TableConfig) bool {
	var str string

	struct_row_name := _upper_first_char(table.Name)
	struct_table_name := struct_row_name + "TableProxy"

	// table
	str += ("type " + struct_table_name + " struct {\n")
	//str += "	db *mysql_proxy.DB\n"
	str += "	tables_mgr *TablesProxyManager\n"
	if table.SingleRow {
		str += "	row *" + struct_row_name + "\n"
	}
	str += "}\n\n"

	// init func
	//str += ("func (this *" + struct_table_name + ") Init(db *mysql_proxy.DB) {\n")
	str += ("func (this *" + struct_table_name + ") Init(tables_mgr *TablesProxyManager) {\n")
	//str += ("	this.db = db\n")
	str += ("	this.tables_mgr = tables_mgr\n")
	str += "}\n\n"

	var field_list string
	for i, field := range table.Fields {
		is_unsigned := strings.Contains(strings.ToLower(field.TypeStr), "unsigned")
		go_type := mysql_base.MysqlFieldType2GoTypeStr(field.Type, is_unsigned)
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
		is_unsigned := strings.Contains(strings.ToLower(field.TypeStr), "unsigned")
		go_type := mysql_base.MysqlFieldType2GoTypeStr(field.Type, is_unsigned)
		if go_type == "" {
			continue
		}

		var dest string
		if field.StructName != "" && (mysql_base.IsMysqlFieldBinaryType(field.Type) || mysql_base.IsMysqlFieldBlobType(field.Type)) {
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
	str += ("	var t = Create" + struct_row_name + "()\n")
	if bytes_define_list != "" {
		str += ("	var " + bytes_define_list + " []byte\n")
	}
	str += ("	var err error\n")
	str += ("	var dest_list = []interface{}{" + dest_list + "}\n")
	if !table.SingleRow {
		//str += ("	err = this.db.Select(\"" + table.Name + "\", field_name, field_value, field_list, dest_list)\n")
		str += ("	err = this.tables_mgr.proxy.Select(this.tables_mgr.host_id, this.tables_mgr.db_name, \"" + table.Name + "\", field_name, field_value, field_list, dest_list)\n")
	} else {
		//str += ("	err = this.db.Select(\"" + table.Name + "\", \"place_hold\", 1, field_list, dest_list)\n")
		str += ("	err = this.tables_mgr.proxy.Select(this.tables_mgr.host_id, this.tables_mgr.db_name, \"" + table.Name + "\", \"place_hold\", 1, field_list, dest_list)\n")
	}
	str += ("	if err != nil {\n")
	str += ("		return nil, err\n")
	str += ("	}\n")
	for _, field := range table.Fields {
		if field.StructName != "" && (mysql_base.IsMysqlFieldBinaryType(field.Type) || mysql_base.IsMysqlFieldBlobType(field.Type)) {
			str += "	t.Unmarshal_" + field.Name + "(data_" + field.Name + ")\n"
		}
	}
	str += ("	return t, nil\n")
	str += ("}\n\n")

	// primary field
	var pf *mysql_base.FieldConfig
	var pt string
	if !table.SingleRow {
		var o bool
		pf, pt, o = get_primary_field_and_type(table)
		if !o {
			return false
		}
	}

	if !table.SingleRow {
		// select records count
		str += "func (this *" + struct_table_name + ") SelectRecordsCount() (count int32, err error) {\n"
		//str += "	return this.db.SelectRecordsCount(\"" + table.Name + "\", \"\", nil)\n"
		str += "	return this.tables_mgr.proxy.SelectRecordsCount(this.tables_mgr.host_id, this.tables_mgr.db_name, \"" + table.Name + "\", \"\", nil)\n"
		str += "}\n\n"

		// select records count by field
		str += "func (this *" + struct_table_name + ") SelectRecordsCountByField(field_name string, field_value interface{}) (count int32, err error) {\n"
		//str += "	return this.db.SelectRecordsCount(\"" + table.Name + "\", field_name, field_value)\n"
		str += "	return this.tables_mgr.proxy.SelectRecordsCount(this.tables_mgr.host_id, this.tables_mgr.db_name, \"" + table.Name + "\", field_name, field_value)\n"
		str += "}\n\n"

		// select records condition
		str += "func (this *" + struct_table_name + ") SelectRecordsCondition(field_name string, field_value interface{}, sel_cond *mysql_base.SelectCondition) ([]*" + struct_row_name + ", error) {\n"
		str += "	var field_list = []string{" + field_list + "}\n"
		str += "	var result_list mysql_proxy.QueryResultList\n"
		//str += "	err := this.db.SelectRecordsCondition(\"" + table.Name + "\", field_name, field_value, sel_cond, field_list, &result_list)\n"
		str += "	err := this.tables_mgr.proxy.SelectRecordsCondition(this.tables_mgr.host_id, this.tables_mgr.db_name, \"" + table.Name + "\", field_name, field_value, sel_cond, field_list, &result_list)\n"
		str += "	if err != nil {\n"
		str += "		return nil, err\n"
		str += "	}\n"
		str += gen_get_proxy_result_list(table, struct_row_name, bytes_define_list, dest_list)
		str += "	return r, nil\n"
		str += "}\n\n"

		// select records map condition
		str += "func (this *" + struct_table_name + ") SelectRecordsMapCondition(field_name string, field_value interface{}, sel_cond *mysql_base.SelectCondition) (map[" + pt + "]*" + struct_row_name + ", error) {\n"
		str += "	var field_list = []string{" + field_list + "}\n"
		//str += "	records_map, err := this.db.SelectRecordsMapCondition(\"" + table.Name + "\", field_name, field_value, sel_cond, field_list)\n"
		str += "	records_map, err := this.tables_mgr.proxy.SelectRecordsMapCondition(this.tables_mgr.host_id, this.tables_mgr.db_name, \"" + table.Name + "\", field_name, field_value, sel_cond, field_list)\n"
		str += "	if err != nil {\n"
		str += "		return nil, err\n"
		str += "	}\n"
		str += gen_get_proxy_result_map(table, struct_row_name, bytes_define_list, dest_list, pt)
		str += "	return r, nil\n"
		str += "}\n\n"

		// select all records
		str += "func (this *" + struct_table_name + ") SelectAllRecords() ([]*" + struct_row_name + ", error) {\n"
		str += "	var field_list = []string{" + field_list + "}\n"
		str += "	var result_list mysql_proxy.QueryResultList\n"
		//str += "	err := this.db.SelectAllRecords(\"" + table.Name + "\", field_list, &result_list)\n"
		str += "	err := this.tables_mgr.proxy.SelectAllRecords(this.tables_mgr.host_id, this.tables_mgr.db_name, \"" + table.Name + "\", field_list, &result_list)\n"
		str += "	if err != nil {\n"
		str += "		return nil, err\n"
		str += "	}\n"
		str += gen_get_proxy_result_list(table, struct_row_name, bytes_define_list, dest_list)
		str += "	return r, nil\n"
		str += "}\n\n"

		// select all records map
		str += "func (this *" + struct_table_name + ") SelectAllRecordsMap() (map[" + pt + "]*" + struct_row_name + ", error) {\n"
		str += "	var field_list = []string{" + field_list + "}\n"
		//str += "	records_map, err := this.db.SelectAllRecordsMap(\"" + table.Name + "\", field_list)\n"
		str += "	records_map, err := this.tables_mgr.proxy.SelectAllRecordsMap(this.tables_mgr.host_id, this.tables_mgr.db_name, \"" + table.Name + "\", field_list)\n"
		str += "	if err != nil {\n"
		str += "		return nil, err\n"
		str += "	}\n"
		str += gen_get_proxy_result_map(table, struct_row_name, bytes_define_list, dest_list, pt)
		str += "	return r, nil\n"
		str += "}\n\n"

		// select records
		str += "func (this *" + struct_table_name + ") SelectRecords(field_name string, field_value interface{}) ([]*" + struct_row_name + ", error) {\n"
		str += "	var field_list = []string{" + field_list + "}\n"
		str += "	var result_list mysql_proxy.QueryResultList\n"
		//str += "	err := this.db.SelectRecords(\"" + table.Name + "\", field_name, field_value, field_list, &result_list)\n"
		str += "	err := this.tables_mgr.proxy.SelectRecords(this.tables_mgr.host_id, this.tables_mgr.db_name, \"" + table.Name + "\", field_name, field_value, field_list, &result_list)\n"
		str += "	if err != nil {\n"
		str += "		return nil, err\n"
		str += "	}\n"
		str += gen_get_proxy_result_list(table, struct_row_name, bytes_define_list, dest_list)
		str += "	return r, nil\n"
		str += "}\n\n"

		// select records map
		str += "func (this *" + struct_table_name + ") SelectRecordsMap(field_name string, field_value interface{}) (map[" + pt + "]*" + struct_row_name + ", error) {\n"
		str += "	var field_list = []string{" + field_list + "}\n"
		//str += "	records_map, err := this.db.SelectRecordsMap(\"" + table.Name + "\", field_name, field_value, field_list)\n"
		str += "	records_map, err := this.tables_mgr.proxy.SelectRecordsMap(this.tables_mgr.host_id, this.tables_mgr.db_name, \"" + table.Name + "\", field_name, field_value, field_list)\n"
		str += "	if err != nil {\n"
		str += "		return nil, err\n"
		str += "	}\n"
		str += gen_get_proxy_result_map(table, struct_row_name, bytes_define_list, dest_list, pt)
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
		//str += ("	dest_list, err := this.db.SelectField(\"" + table.Name + "\", \"" + pf.Name + "\")\n")
		str += ("	dest_list, err := this.tables_mgr.proxy.SelectField(this.tables_mgr.host_id, this.tables_mgr.db_name, \"" + table.Name + "\", \"" + pf.Name + "\")\n")
		str += ("	if err != nil {\n")
		str += ("		return nil, err\n")
		str += ("	}\n")
		str += ("	var fields =  make([]" + pt + ", len(dest_list))\n")
		str += ("	for i:=0; i<len(dest_list); i++ {\n")
		str += ("		fields[i] = dest_list[i].(" + pt + ")\n")
		str += ("	}\n")
		str += ("	return fields, nil\n")
		str += ("}\n\n")

		// select all primary field map
		str += ("func (this *" + struct_table_name + ") SelectAllPrimaryFieldMap() (map[" + pt + "]bool, error) {\n")
		//str += ("	dest_list, err := this.db.SelectField(\"" + table.Name + "\", \"" + pf.Name + "\")\n")
		str += ("	dest_list, err := this.tables_mgr.proxy.SelectField(this.tables_mgr.host_id, this.tables_mgr.db_name, \"" + table.Name + "\", \"" + pf.Name + "\")\n")
		str += ("	if err != nil {\n")
		str += ("		return nil, err\n")
		str += ("	}\n")
		str += ("	var fields_map =  make(map[" + pt + "]bool, len(dest_list))\n")
		str += ("	for i:=0; i<len(dest_list); i++ {\n")
		str += ("		fields_map[dest_list[i].(" + pt + ")] = true\n")
		str += ("	}\n")
		str += ("	return fields_map, nil\n")
		str += ("}\n\n")

		// insert
		str += "func (this *" + struct_table_name + ") Insert(t *" + struct_row_name + ") {\n"
		str += "	var field_list = t._format_field_list()\n"
		str += "	if field_list != nil {\n"
		//str += "		this.db.Insert(\"" + table.Name + "\", field_list)\n"
		str += "		this.tables_mgr.proxy.Insert(this.tables_mgr.host_id, this.tables_mgr.db_name, \"" + table.Name + "\", field_list)\n"
		str += "	}\n"
		str += "}\n\n"

		// insert ignore
		str += "func (this *" + struct_table_name + ") InsertIgnore(t *" + struct_row_name + ") {\n"
		str += "	var field_list = t._format_field_list()\n"
		str += "	if field_list != nil {\n"
		//str += "		this.db.InsertIgnore(\"" + table.Name + "\", field_list)\n"
		str += "		this.tables_mgr.proxy.InsertIgnore(this.tables_mgr.host_id, this.tables_mgr.db_name, \"" + table.Name + "\", field_list)\n"
		str += "	}\n"
		str += "}\n\n"

		// delete
		str += ("func (this *" + struct_table_name + ") Delete(" + pf.Name + " " + pt + ") {\n")
		//str += ("	this.db.Delete(\"" + table.Name + "\", \"" + pf.Name + "\", " + pf.Name + ")\n")
		str += ("	this.tables_mgr.proxy.Delete(this.tables_mgr.host_id, this.tables_mgr.db_name, \"" + table.Name + "\", \"" + pf.Name + "\", " + pf.Name + ")\n")
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
		//str += "		this.db.Update(\"" + table.Name + "\", \"" + pf.Name + "\", t.Get_" + pf.Name + "(), field_list)\n"
		str += "		this.tables_mgr.proxy.Update(this.tables_mgr.host_id, this.tables_mgr.db_name, \"" + table.Name + "\", \"" + pf.Name + "\", t.Get_" + pf.Name + "(), field_list)\n"
	} else {
		//str += "		this.db.Update(\"" + table.Name + "\", \"place_hold\", 1, field_list)\n"
		str += "		this.tables_mgr.proxy.Update(this.tables_mgr.host_id, this.tables_mgr.db_name, \"" + table.Name + "\", \"place_hold\", 1, field_list)\n"
	}
	str += "	}\n"
	str += "}\n\n"

	// update some field
	if !table.SingleRow {
		str += "func (this *" + struct_table_name + ") UpdateWithFVPList(" + pf.Name + " " + pt + ", field_list []*mysql_base.FieldValuePair) {\n"
		//str += "	this.db.Update(\"" + table.Name + "\", \"" + pf.Name + "\", " + pf.Name + ", field_list)\n"
		str += "	this.tables_mgr.proxy.Update(this.tables_mgr.host_id, this.tables_mgr.db_name, \"" + table.Name + "\", \"" + pf.Name + "\", " + pf.Name + ", field_list)\n"
	} else {
		str += "func (this *" + struct_table_name + ") UpdateWithFVPList(field_list []*mysql_base.FieldValuePair) {\n"
		//str += "	this.db.Update(\"" + table.Name + "\", \"place_hold\", 1, field_list)\n"
		str += "	this.tables_mgr.proxy.Update(this.tables_mgr.host_id, this.tables_mgr.db_name, \"" + table.Name + "\", \"place_hold\", 1, field_list)\n"
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

	str += gen_procedure_proxy_source(table, struct_table_name, struct_row_name, pf, pt)

	_, err := f.WriteString(str)
	if err != nil {
		log.Infof("write string err %v", err.Error())
		return false
	}

	return true
}

func gen_procedure_proxy_source(table *mysql_base.TableConfig, struct_table_name, struct_row_name string, primary_field *mysql_base.FieldConfig, primary_type string) string {
	var str string

	if !table.SingleRow {
		str += "func (this *" + struct_table_name + ") TransactionInsert(transaction *mysql_proxy.Transaction, t *" + struct_row_name + ") {\n"
		str += "	field_list := t._format_field_list()\n"
		str += "	if field_list != nil {\n"
		str += "		transaction.Insert(\"" + table.Name + "\", field_list)\n"
		str += "	}\n"
		str += "}\n\n"
		str += "func (this *" + struct_table_name + ") TransactionDelete(transaction *mysql_proxy.Transaction, " + primary_field.Name + " " + primary_type + ") {\n"
		str += "	transaction.Delete(\"" + table.Name + "\", \"" + primary_field.Name + "\", " + primary_field.Name + ")\n"
		str += "}\n\n"
	}

	str += "func (this *" + struct_table_name + ") TransactionUpdateAll(transaction *mysql_proxy.Transaction, t*" + struct_row_name + ") {\n"
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
		str += "func (this *" + struct_table_name + ") TransactionUpdateWithFVPList(transaction *mysql_proxy.Transaction, " + primary_field.Name + " " + primary_type + ", field_list []*mysql_base.FieldValuePair) {\n"
		str += "	transaction.Update(\"" + table.Name + "\", \"" + primary_field.Name + "\", " + primary_field.Name + ", field_list)\n"
	} else {
		str += "func (this *" + struct_table_name + ") TransactionUpdateWithFVPList(transaction *mysql_proxy.Transaction, field_list []*mysql_base.FieldValuePair) {\n"
		str += "	transaction.Update(\"" + table.Name + "\", \"place_hold\", 1, field_list)\n"
	}
	str += "}\n\n"

	str += "func (this *" + struct_table_name + ") TransactionUpdateWithFieldName(transaction *mysql_proxy.Transaction, t *" + struct_row_name + ", fields_name []string) {\n"
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
