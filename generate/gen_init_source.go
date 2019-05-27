package mysql_generate

import (
	"log"
	"os"

	"github.com/huoshan017/mysql-go/base"
)

func gen_init_source(f *os.File, pkg_name string, tables []*mysql_base.TableConfig) bool {
	var str string
	str += "package " + pkg_name + "\n\n"
	str += "import (\n"
	str += "	\"github.com/huoshan017/mysql-go/manager\"\n"
	str += "	\"github.com/huoshan017/mysql-go/proxy/client\"\n"
	str += ")\n\n"

	str += "type TablesManager struct {\n"
	for _, table := range tables {
		var struct_row_name = _upper_first_char(table.Name)
		var struct_table_name = struct_row_name + "_Table"
		var table_instance = "db_" + struct_table_name

		str += "	" + table_instance + " *" + struct_table_name + "\n"
	}
	str += "}\n\n"

	str += "func (this *TablesManager) Init(db *mysql_manager.DB) {\n"
	for _, table := range tables {
		var struct_row_name = _upper_first_char(table.Name)
		var struct_table_name = struct_row_name + "_Table"
		var table_instance = "db_" + struct_table_name

		str += "	this." + table_instance + " = &" + struct_table_name + "{}\n"
		str += "	this." + table_instance + ".Init(db)\n"
	}
	str += "}\n\n"

	for _, table := range tables {
		var struct_row_name = _upper_first_char(table.Name)
		var struct_table_name = struct_row_name + "_Table"
		var table_instance = "db_" + struct_table_name

		str += "func (this *TablesManager) Get_" + struct_table_name + "() *" + struct_table_name + "{\n"
		str += "	return this." + table_instance + "\n"
		str += "}\n\n"
	}

	str += "func NewTablesManager(db *mysql_manager.DB) *TablesManager {\n"
	str += "	tm := &TablesManager{}\n"
	str += "	tm.Init(db)\n"
	str += "	return tm\n"
	str += "}\n\n"

	// proxy following
	str += "type TableProxysManager struct {\n"
	for _, table := range tables {
		var struct_row_name = _upper_first_char(table.Name)
		var struct_table_name = struct_row_name + "_Table_Proxy"
		var table_instance = "db_" + struct_table_name

		str += "	" + table_instance + " *" + struct_table_name + "\n"
	}
	str += "}\n\n"

	str += "func (this *TableProxysManager) Init(db *mysql_proxy.DB) {\n"
	for _, table := range tables {
		var struct_row_name = _upper_first_char(table.Name)
		var struct_table_name = struct_row_name + "_Table_Proxy"
		var table_instance = "db_" + struct_table_name

		str += "	this." + table_instance + " = &" + struct_table_name + "{}\n"
		str += "	this." + table_instance + ".Init(db)\n"
	}
	str += "}\n\n"

	for _, table := range tables {
		var struct_row_name = _upper_first_char(table.Name)
		var struct_table_name = struct_row_name + "_Table_Proxy"
		var table_instance = "db_" + struct_table_name

		str += "func (this *TableProxysManager) Get_" + struct_table_name + "() *" + struct_table_name + "{\n"
		str += "	return this." + table_instance + "\n"
		str += "}\n\n"
	}

	str += "func NewTableProxysManager(db *mysql_proxy.DB) *TableProxysManager {\n"
	str += "	tm := &TableProxysManager{}\n"
	str += "	tm.Init(db)\n"
	str += "	return tm\n"
	str += "}\n"

	_, err := f.WriteString(str)
	if err != nil {
		log.Printf("write string err %v\n", err.Error())
		return false
	}

	return true
}
