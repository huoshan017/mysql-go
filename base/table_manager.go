package mysql_base

import (
	"log"
	"sync"
)

const (
	DB_OPERATE_TYPE_NONE          = iota
	DB_OPERATE_TYPE_INSERT_RECORD = 1
	DB_OPERATE_TYPE_DELETE_RECORD = 2
	DB_OPERATE_TYPE_UPDATE_RECORD = 3
)

type TableBase interface {
	Insert() bool
	Update() bool
	Delete() bool
	Name() string
}

type OpData struct {
	table_name string
	op_type    int32
	field_list []*FieldValuePair
}

type TableOpData struct {
	records_ops map[interface{}]*OpData
	locker      sync.RWMutex
}

func (this *TableOpData) Init(table_name string) {
	this.locker.Lock()
	defer this.locker.Unlock()
	this.records_ops = make(map[interface{}]*OpData)
}

func (this *TableOpData) insert(key interface{}, field_args ...FieldValuePair) {
	var op_data *OpData
	this.locker.RLock()
	var o bool
	op_data, o = this.records_ops[key]
	this.locker.RUnlock()

	if !o {
		this.locker.Lock()
		op_data = &OpData{}
		_, o = this.records_ops[key]
		if !o {
			this.records_ops[key] = op_data
		}
		this.locker.Unlock()
	}

	if op_data.op_type == DB_OPERATE_TYPE_NONE {

	}
}

func (this *TableOpData) update(key interface{}, field_args ...FieldValuePair) {

}

type TablesOpMgr struct {
	op_list           *List
	table_records_ops map[string]*TableOpData
}

func (this *TablesOpMgr) Init() {
	this.op_list = &List{}
	this.table_records_ops = make(map[string]*TableOpData)
}

type TableManager struct {
	tables_map map[string]TableBase
	locker     sync.RWMutex
	db         *Database
}

func (this *TableManager) Init(db *Database) {
	this.tables_map = make(map[string]TableBase)
	this.db = db
}

func (this *TableManager) Add(table TableBase) bool {
	name := table.Name()
	var t TableBase

	this.locker.Lock()
	defer this.locker.Unlock()

	t = this.tables_map[name]
	if t != nil {
		log.Printf("TableManager::Add already has table %v", name)
		return false
	}
	this.tables_map[name] = table
	return true
}

func (this *TableManager) RemoveByName(name string) bool {
	this.locker.Lock()
	defer this.locker.Unlock()

	if this.tables_map == nil {
		return false
	}
	if _, o := this.tables_map[name]; !o {
		return false
	}
	delete(this.tables_map, name)
	return true
}

func (this *TableManager) Remove(table TableBase) bool {
	return this.RemoveByName(table.Name())
}

func (this *TableManager) Get(name string) TableBase {
	this.locker.RLock()
	defer this.locker.RUnlock()

	if this.tables_map == nil {
		return nil
	}
	return this.tables_map[name]
}

func (this *TableManager) InsertRecord(table_name string, field_args ...FieldValuePair) bool {
	table := this.Get(table_name)
	if table == nil {
		return false
	}

	/*var field_list []*FieldValuePair
	for _, f := range field_args {
		field_list = append(field_list, &f)
	}

	this.op_list = append(this.op_list, &OpData{
		op_type:    DB_OPERATE_TYPE_INSERT_RECORD,
		table_name: table_name,
		field_list: field_list,
	})*/
	return true
}

func (this *TableManager) DeleteRecord(table_name, field_name string, field_value interface{}) bool {
	table := this.Get(table_name)
	if table == nil {
		return false
	}

	/*fp := &FieldValuePair{
		Name:  field_name,
		Value: field_value,
	}

	this.op_list = append(this.op_list, &OpData{
		op_type:    DB_OPERATE_TYPE_DELETE_RECORD,
		table_name: table_name,
		field_list: []*FieldValuePair{fp},
	})*/

	return true
}

func (this *TableManager) UpdateTableField(table_name, field_name string, field_value interface{}) bool {
	return true
}
