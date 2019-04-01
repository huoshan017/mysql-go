package mysql_generator

import (
	//"log"
	"os"
	"path/filepath"

	"github.com/huoshan017/mysql-go/base"
)

func gen_source(f *os.File, dest_dir string, table *mysql_base.TableConfig) (err error) {
	_, pkg := filepath.Split(dest_dir)
	str := "package " + pkg + "\n\nimport (\n"
	str += "	\"encoding/csv\"\n"
	str += "	\"io/ioutil\"\n"
	str += "	\"log\"\n"
	str += "	\"strconv\"\n"
	str += "	\"strings\"\n"
	str += ")\n\n"

	// struct
	str += ("type " + table.Name + " struct {\n")
	for _, f := range table.Fields {

	}
	str += "}\n\n"
	return
}
