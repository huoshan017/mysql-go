package mysql_generate

import (
	"os"

	mysql_base "github.com/huoshan017/mysql-go/base"
	"github.com/huoshan017/mysql-go/log"
)

func gen_init_source(f *os.File, pkg_name string, configBytes []byte, tables []*mysql_base.TableConfig) bool {
	var str string
	str += "package " + pkg_name + "\n\n"
	str += "import (\n"
	str += "	\"fmt\"\n"
	str += "	\"github.com/huoshan017/mysql-go/manager\"\n"
	str += "	\"github.com/huoshan017/mysql-go/proxy/client\"\n"
	str += "	\"github.com/huoshan017/mysql-go/generate\"\n"
	str += ")\n\n"

	str += "var configBytes = `" + string(configBytes) + "`\n\n"

	str += "var configLoader mysql_generate.ConfigLoader" + "\n\n"

	str += "type TablesManager struct {\n"
	for _, table := range tables {
		var struct_row_name = _upper_first_char(table.Name)
		var struct_table_name = struct_row_name + "Table"
		var table_instance = "db" + struct_table_name

		str += "	" + table_instance + " *" + struct_table_name + "\n"
	}
	str += "}\n\n"

	str += "func (this *TablesManager) Init(db *mysql_manager.DB) {\n"
	str += "	if !configLoader.LoadConfigBytes([]byte(configBytes)) {\n"
	str += "		panic(\"load config bytes crash\")\n"
	str += "	}\n\n"
	str += "	if err := db.SetConfigLoader(&configLoader); err != nil {\n"
	str += "		panic(fmt.Sprintf(\"set config loader err: %v\", err))\n"
	str += "	}\n\n"
	for _, table := range tables {
		var struct_row_name = _upper_first_char(table.Name)
		var struct_table_name = struct_row_name + "Table"
		var table_instance = "db" + struct_table_name

		str += "	this." + table_instance + " = &" + struct_table_name + "{}\n"
		str += "	this." + table_instance + ".Init(db)\n"
	}
	str += "}\n\n"

	for _, table := range tables {
		var struct_row_name = _upper_first_char(table.Name)
		var struct_table_name = struct_row_name + "Table"
		var table_instance = "db" + struct_table_name

		str += "func (this *TablesManager) Get" + struct_table_name + "() *" + struct_table_name + "{\n"
		str += "	return this." + table_instance + "\n"
		str += "}\n\n"
	}

	str += "func NewTablesManager(db *mysql_manager.DB) *TablesManager {\n"
	str += "	tm := &TablesManager{}\n"
	str += "	tm.Init(db)\n"
	str += "	return tm\n"
	str += "}\n\n"

	// proxy following
	str += "type TablesProxyManager struct {\n"
	str += "	proxy *mysql_proxy.DB\n"
	str += "	host_id int32\n"
	str += "	db_name string\n"
	for _, table := range tables {
		var struct_row_name = _upper_first_char(table.Name)
		var struct_table_name = struct_row_name + "TableProxy"
		var table_instance = "db" + struct_table_name

		str += "	" + table_instance + " *" + struct_table_name + "\n"
	}
	str += "}\n\n"

	str += "func (this *TablesProxyManager) Init(proxy *mysql_proxy.DB, host_id int32, db_name string) {\n"
	str += "	this.proxy = proxy\n"
	str += "	this.host_id = host_id\n"
	str += "	this.db_name = db_name\n"
	for _, table := range tables {
		var struct_row_name = _upper_first_char(table.Name)
		var struct_table_name = struct_row_name + "TableProxy"
		var table_instance = "db" + struct_table_name

		str += "	this." + table_instance + " = &" + struct_table_name + "{}\n"
		str += "	this." + table_instance + ".Init(this)\n"
	}
	str += "}\n\n"

	for _, table := range tables {
		var struct_row_name = _upper_first_char(table.Name)
		var struct_table_name = struct_row_name + "TableProxy"
		var table_instance = "db" + struct_table_name

		str += "func (this *TablesProxyManager) Get" + struct_table_name + "() *" + struct_table_name + "{\n"
		str += "	return this." + table_instance + "\n"
		str += "}\n\n"
	}

	str += "func NewTablesProxyManager(proxy *mysql_proxy.DB, host_id int32, db_name string) *TablesProxyManager {\n"
	str += "	tm := &TablesProxyManager{}\n"
	str += "	tm.Init(proxy, host_id, db_name)\n"
	str += "	return tm\n"
	str += "}\n"

	_, err := f.WriteString(str)
	if err != nil {
		log.Infof("write string err %v", err.Error())
		return false
	}

	return true
}
