package mysql_generate

import (
	"log"
	"os"

	//"strconv"
	//"strings"

	"github.com/huoshan017/mysql-go/base"
)

func gen_record_mgr_source(f *os.File, pkg_name string, table *mysql_base.TableConfig) bool {
	if table.SingleRow {
		return true
	}

	// primary field
	pf, pt, o := get_primary_field_and_type(table)
	if !o {
		return false
	}

	struct_row_name := _upper_first_char(table.Name)
	record_mgr_name := struct_row_name + "RecordMgr"

	str := "type " + record_mgr_name + " struct {\n"
	str += "	load_list *simplelru.LRU\n"
	str += "	locker sync.RWMutex\n"
	str += "}\n\n"

	str += "func New" + struct_row_name + "RecordMgr(record_count int) *" + record_mgr_name + "{\n"
	str += "	lists, err := simplelru.NewLRU(record_count, nil)\n"
	str += "	if err != nil {\n"
	str += "		return nil\n"
	str += "	}\n"
	str += "	return &" + record_mgr_name + "{\n"
	str += "		load_list: lists,\n"
	str += "	}\n"
	str += "}\n\n"

	str += "func (this *" + record_mgr_name + ") Add(new_row *" + struct_row_name + ") bool {\n"
	str += "	this.locker.Lock()\n"
	str += "	defer this.locker.Unlock()\n"
	str += "	key := new_row." + pf.Name + "\n"
	str += "	if this.load_list.Contains(key) {\n"
	str += "		return false\n"
	str += "	}\n"
	str += "	this.load_list.Add(key, new_row)\n"
	str += "	return true\n"
	str += "}\n\n"

	str += "func (this *" + record_mgr_name + ") Has(key " + pt + ") bool {\n"
	str += "	this.locker.RLock()\n"
	str += "	defer this.locker.RUnlock()\n"
	str += "	return this.load_list.Contains(key)\n"
	str += "}\n\n"

	str += "func (this *" + record_mgr_name + ") Get(key " + pt + ") *" + struct_row_name + " {\n"
	str += "	this.locker.RLock()\n"
	str += "	defer this.locker.RUnlock()\n"
	str += "	d, o := this.load_list.Get(key)\n"
	str += "	if !o {\n"
	str += "		return nil\n"
	str += "	}\n"
	str += "	return d.(*" + struct_row_name + ")\n"
	str += "}\n\n"

	str += "func (this *" + record_mgr_name + ") Remove(key " + pt + ") bool {\n"
	str += "	this.locker.Lock()\n"
	str += "	defer this.locker.Unlock()\n"
	str += "	return this.load_list.Remove(key)\n"
	str += "}\n\n"

	_, err := f.WriteString(str)
	if err != nil {
		log.Printf("write string err %v\n", err.Error())
		return false
	}

	return true
}
