package mysql_generate

import (
	"log"
	"os"

	//"strconv"
	//"strings"

	"github.com/huoshan017/mysql-go/base"
)

func gen_record_mgr_source(f *os.File, pkg_name string, table *mysql_base.TableConfig) bool {
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

	struct_row_name := _upper_first_char(table.Name)
	record_mgr_name := struct_row_name + "_RecordMgr"

	str := "type " + record_mgr_name + " struct {\n"
	str += "	load_list *simplelru.LRU\n"
	str += "	have_map map[" + pt + "]bool\n"
	str += "	locker sync.RWMutex\n"
	str += "}\n\n"

	str += "func New_" + struct_row_name + "_RecordMgr(record_count int) *" + record_mgr_name + "{\n"
	str += "	lists, err := simplelru.NewLRU(record_count, nil)\n"
	str += "	if err != nil {\n"
	str += "		return nil\n"
	str += "	}\n"
	str += "	return &" + record_mgr_name + "{\n"
	str += "		load_list: lists,\n"
	str += "		have_map:  make(map[" + pt + "]bool),\n"
	str += "	}\n"
	str += "}\n\n"

	str += "func (this *" + record_mgr_name + ") LoadRecordsWith() {\n"
	str += "}\n\n"

	str += "func (this *" + record_mgr_name + ") New(key " + pt + ") *" + struct_row_name + " {\n"
	str += "	this.locker.Lock()\n"
	str += "	defer this.locker.Unlock()\n"
	str += "	if this.load_list.Contains(key) {\n"
	str += "		return nil\n"
	str += "	}\n"
	str += "	d := &" + struct_row_name + "{\n"
	str += "		" + pf.Name + ": key,\n"
	str += "	}\n"
	str += "	this.load_list.Add(key, d)\n"
	str += "	return d\n"
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
