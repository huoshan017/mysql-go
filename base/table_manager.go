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
	op_type    int32
	table_name string
	field_list []*FieldValuePair
	next       *OpData
	prev       *OpData
}

type MapOpData struct {
	records_ops map[interface{}]*OpData
	locker      sync.RWMutex
}

func (this *MapOpData) get(key interface{}) *OpData {
	this.locker.RLock()
	defer this.locker.RUnlock()
	if this.records_ops == nil {
		return nil
	}
	return this.records_ops[key]
}

func (this *MapOpData) insert(key interface{}, field_args ...FieldValuePair) {
	var op_data *OpData
	this.locker.RLock()
	if this.records_ops == nil {
		this.locker.RUnlock()
		this.locker.Lock()
		if this.records_ops == nil {
			this.records_ops = make(map[interface{}]*OpData)
		}
		op_data = this.records_ops[key]
		if op_data == nil {
			op_data = &OpData{
				op_type: DB_OPERATE_TYPE_INSERT_RECORD,
			}
			this.records_ops[key] = op_data
		}
		this.locker.Unlock()
	}

	if op_data.op_type == DB_OPERATE_TYPE_DELETE_RECORD {

	}

	this.locker.RUnlock()
}

type TableManager struct {
	tables_map        map[string]TableBase
	locker            sync.RWMutex
	db                *Database
	op_list           List
	table_records_ops map[string]*MapOpData
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
