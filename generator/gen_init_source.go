package mysql_generator

import (
	"log"
	"os"
)

func gen_init_source(f *os.File, pkg_name string) bool {
	var str string
	str += "package " + pkg_name + "\n"
	str += "import (\n"
	str += "	\"github.com/huoshan017/mysql-go/manager\"\n"
	str += ")\n\n"
	str += "var db_mgr *mysql_manager.DB\n\n"
	str += "func SetDB(db *mysql_manager.DB) {\n"
	str += "	db_mgr = db\n"
	str += "}\n\n"
	str += "func GetDB() *mysql_manager.DB {\n"
	str += "	return db_mgr\n"
	str += "}\n\n"

	_, err := f.WriteString(str)
	if err != nil {
		log.Printf("write string err %v\n", err.Error())
		return false
	}

	return true
}
