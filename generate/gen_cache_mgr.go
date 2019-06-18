package mysql_generate

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/huoshan017/mysql-go/base"
)

func gen_cache_mgr_source(f *os.File, pkg_name string, table *mysql_base.TableConfig) string {
	var str string

	struct_row_name := _upper_first_char(table.Name)
	cache_mgr_name := struct_row_name + "CacheMgr"

	return str
}
