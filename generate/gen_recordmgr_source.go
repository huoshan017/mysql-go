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
	select_func_name := struct_row_name + "SelectFunc"

	str := "type " + record_mgr_name + " struct {\n"
	str += "	load_list *simplelru.LRU\n"
	str += "	have_map map[" + pt + "]bool\n"
	str += "	locker sync.RWMutex\n"
	str += "	select_func " + select_func_name + "\n"
	str += "}\n\n"

	str += "type " + select_func_name + " func() (*" + struct_row_name + ", error)\n\n"

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

	str += "func (this *" + record_mgr_name + ") SetSelectFunc(sel_func " + select_func_name + ") {\n"
	str += "	this.select_func = sel_func\n"
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

	str += "func (this *" + record_mgr_name + ") Get(key " + pt + ", is_sel bool) *" + struct_row_name + " {\n"
	str += "	this.locker.RLock()\n"
	str += "	defer this.locker.RUnlock()\n"
	str += "	d, o := this.load_list.Get(key)\n"
	str += "	if !o {\n"
	str += "		if !is_sel {\n"
	str += "			return nil\n"
	str += "		}\n"
	str += "		if this.select_func == nil {\n"
	str += "			return nil\n"
	str += "		}\n"
	str += "		sel_row, err := this.select_func()\n"
	str += "		if err != nil {\n"
	str += "			return nil\n"
	str += "		}\n"
	str += "		this.load_list.Add(key, sel_row)\n"
	str += "		return sel_row\n"
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
