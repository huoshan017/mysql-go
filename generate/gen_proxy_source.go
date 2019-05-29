package mysql_generate

import (
	"log"
	"os"
	"strings"

	"github.com/huoshan017/mysql-go/base"
)

func gen_proxy_source(f *os.File, pkg_name string, table *mysql_base.TableConfig) bool {
	struct_row_name := _upper_first_char(table.Name)
	struct_table_name := struct_row_name + "_Table_Proxy"

	var str string

	// table
	str += ("type " + struct_table_name + " struct {\n")
	str += "	db *mysql_proxy.DB\n"
	if table.SingleRow {
		str += "	row *" + struct_row_name + "\n"
	}
	str += "}\n\n"

	// init func
	str += ("func (this *" + struct_table_name + ") Init(db *mysql_proxy.DB) {\n")
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
		str += ("func (this *" + struct_table_name + ") Select(field_name string, field_value interface{}) (*" + struct_row_name + ", bool) {\n")
	} else {
		str += "func (this *" + struct_table_name + ") Select() (*" + struct_row_name + ", bool) {\n"
	}
	str += ("	var field_list = []string{" + field_list + "}\n")
	str += ("	var t = Create_" + struct_row_name + "()\n")
	if bytes_define_list != "" {
		str += ("	var " + bytes_define_list + " []byte\n")
	}
	str += ("	var dest_list = []interface{}{" + dest_list + "}\n")
	if !table.SingleRow {
		str += ("	if !this.db.Select(\"" + table.Name + "\", field_name, field_value, field_list, dest_list) {\n")
	} else {
		str += ("	if !this.db.Select(\"" + table.Name + "\", \"place_hold\", 1, field_list, dest_list) {\n")
	}
	str += ("		return nil, false\n")
	str += ("	}\n")
	for _, field := range table.Fields {
		if field.StructName != "" && (mysql_base.IsMysqlFieldBinaryType(field.RealType) || mysql_base.IsMysqlFieldBlobType(field.RealType)) {
			str += "	t.Unmarshal_" + field.Name + "(data_" + field.Name + ")\n"
		}
	}
	str += ("	return t, true\n")
	str += ("}\n\n")

	if !table.SingleRow {
		// select records condition
		str += "func (this *" + struct_table_name + ") SelectRecordsCondition(field_name string, field_value interface{}, sel_cond *mysql_base.SelectCondition) ([]*" + struct_row_name + ", bool) {\n"
		str += "	var field_list = []string{" + field_list + "}\n"
		str += "	var result_list mysql_proxy.QueryResultList\n"
		str += "	if !this.db.SelectRecordsCondition(\"" + table.Name + "\", field_name, field_value, sel_cond, field_list, &result_list) {\n"
		str += "		return nil, false\n"
		str += "	}\n"
		str += gen_get_result_list(table, struct_row_name, bytes_define_list, dest_list)
		str += "	return r, true\n"
		str += "}\n\n"

		// select all records
		str += "func (this *" + struct_table_name + ") SelectAllRecords() ([]*" + struct_row_name + ", bool) {\n"
		str += "	var field_list = []string{" + field_list + "}\n"
		str += "	var result_list mysql_proxy.QueryResultList\n"
		str += "	if !this.db.SelectAllRecords(\"" + table.Name + "\", field_list, &result_list) {\n"
		str += "		return nil, false\n"
		str += "	}\n"
		str += gen_get_result_list(table, struct_row_name, bytes_define_list, dest_list)
		str += "	return r, true\n"
		str += "}\n\n"

		// select records
		str += "func (this *" + struct_table_name + ") SelectRecords(field_name string, field_value interface{}) ([]*" + struct_row_name + ", bool) {\n"
		str += "	var field_list = []string{" + field_list + "}\n"
		str += "	var result_list mysql_proxy.QueryResultList\n"
		str += "	if !this.db.SelectRecords(\"" + table.Name + "\", field_name, field_value, field_list, &result_list) {\n"
		str += "		return nil, false\n"
		str += "	}\n"
		str += gen_get_result_list(table, struct_row_name, bytes_define_list, dest_list)
		str += "	return r, true\n"
		str += "}\n\n"
	}

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
		// select primary field
		str += ("func (this *" + struct_table_name + ") SelectByPrimaryField(key " + pt + ") *" + struct_row_name + " {\n")
		str += ("	v, o := this.Select(\"" + pf.Name + "\", key)\n")
		str += ("	if !o {\n")
		str += ("		return nil\n")
		str += ("	}\n")
		str += ("	return v\n")
		str += ("}\n\n")

		// select all primary field
		str += ("func (this *" + struct_table_name + ") SelectAllPrimaryField() ([]" + pt + ") {\n")
		str += ("	dest_list, o := this.db.SelectField(\"" + table.Name + "\", \"" + pf.Name + "\")\n")
		str += ("	if !o {\n")
		str += ("		return nil\n")
		str += ("	}\n")
		str += ("	var fields =  make([]" + pt + ", len(dest_list))\n")
		str += ("	for i:=0; i<len(dest_list); i++ {\n")
		str += ("		fields[i] = dest_list[i].(" + pt + ")\n")
		str += ("	}\n")
		str += ("	return fields\n")
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
		str += "func (this *" + struct_table_name + ") NewRow(" + pf.Name + " " + pt + ") *" + struct_row_name + " {\n"
		str += "	return &" + struct_row_name + "{ " + pf.Name + ": " + pf.Name + ", }\n"
		str += "}\n\n"
	} else {
		str += "func (this *" + struct_table_name + ") GetRow() *" + struct_row_name + " {\n"
		str += "	if this.row == nil {\n"
		str += "		row, o := this.Select()\n"
		str += "		if !o {\n"
		str += "			return nil\n"
		str += "		}\n"
		str += "		this.row = row\n"
		str += "	}\n"
		str += "	return this.row\n"
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

	str += gen_procedure_proxy_source(table, struct_table_name, struct_row_name, pf, pt)

	_, err := f.WriteString(str)
	if err != nil {
		log.Printf("write string err %v\n", err.Error())
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
