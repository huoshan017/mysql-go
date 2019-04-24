package mysql_generator

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
	str += ")\n\n"

	for _, table := range tables {
		var struct_row_name = _upper_first_char(table.Name)
		var struct_table_name = struct_row_name + "_Table"

		var table_instance = "db_" + struct_table_name
		str += "var " + table_instance + " *" + struct_table_name + "\n\n"
		str += "func " + table.Name + "_init(db *mysql_manager.DB) {\n"
		str += "	" + table_instance + " = &" + struct_table_name + "{}\n"
		str += "	" + table_instance + ".Init(db)\n"
		str += "}\n\n"

		str += "func Get_" + struct_table_name + "() *" + struct_table_name + "{\n"
		str += "	return " + table_instance + "\n"
		str += "}\n\n"
	}

	str += "func Init(db *mysql_manager.DB) {\n"
	for _, table := range tables {
		str += "	" + table.Name + "_init(db)\n"
	}
	str += "}\n"

	_, err := f.WriteString(str)
	if err != nil {
		log.Printf("write string err %v\n", err.Error())
		return false
	}

	return true
}
