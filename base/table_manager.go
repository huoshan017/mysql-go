package mysql_base

import (
	//"log"
	"sync"
)

const (
	DB_SQL_TYPE_COMMAND   = iota
	DB_SQL_TYPE_PROCEDURE = 1
)

const (
	DB_OPERATE_TYPE_SELECT = iota
	DB_OPERATE_TYPE_INSERT = 1
	DB_OPERATE_TYPE_DELETE = 2
	DB_OPERATE_TYPE_UPDATE = 3
)

type OpDetail struct {
	table_name string
	op_type    int32
	key        string
	value      interface{}
	field_list []*FieldValuePair
}

type OpData struct {
	sql_type    int32
	detail_list []*OpDetail
}

type OpProcedure struct {
	detail_list []*OpDetail
}

func (this *OpProcedure) Insert() {

}

func (this *OpProcedure) Update() {

}

func (this *OpProcedure) Delete() {

}

type DBOperateManager struct {
	op_list *List
	locker  sync.RWMutex
	db      *Database
}

func (this *DBOperateManager) Init(db *Database) {
	this.op_list = &List{}
	this.db = db
}

func (this *DBOperateManager) Insert(table_name string, field_list []*FieldValuePair) {
	this.locker.Lock()
	defer this.locker.Unlock()

	this.op_list.Append(&OpData{
		sql_type: DB_SQL_TYPE_COMMAND,
		detail_list: []*OpDetail{
			&OpDetail{
				table_name: table_name,
				op_type:    DB_OPERATE_TYPE_INSERT,
				field_list: field_list,
			},
		},
	})
}

func (this *DBOperateManager) Delete(table_name string, field_name string, field_value interface{}) {
	this.locker.Lock()
	defer this.locker.Unlock()

	this.op_list.Append(&OpData{
		sql_type: DB_SQL_TYPE_COMMAND,
		detail_list: []*OpDetail{
			&OpDetail{
				table_name: table_name,
				op_type:    DB_OPERATE_TYPE_DELETE,
				key:        field_name,
				value:      field_value,
			},
		},
	})
}

func (this *DBOperateManager) Update(table_name string, key string, value interface{}, field_list []*FieldValuePair) {
	this.locker.Lock()
	defer this.locker.Unlock()

	this.op_list.Append(&OpData{
		sql_type: DB_SQL_TYPE_COMMAND,
		detail_list: []*OpDetail{
			&OpDetail{
				table_name: table_name,
				op_type:    DB_OPERATE_TYPE_UPDATE,
				key:        key,
				value:      value,
				field_list: field_list,
			},
		},
	})
}

func (this *DBOperateManager) AppendProcedure(procedure *OpProcedure) {
	this.locker.Lock()
	defer this.locker.Unlock()

	this.op_list.Append(&OpData{
		sql_type:    DB_SQL_TYPE_PROCEDURE,
		detail_list: procedure.detail_list,
	})
}

func (this *DBOperateManager) _op_cmd(d *OpDetail) {
	switch d.op_type {
	case DB_OPERATE_TYPE_INSERT:
		{
			this.db.InsertRecord(d.table_name, d.field_list...)
		}
	case DB_OPERATE_TYPE_DELETE:
		{
			this.db.DeleteRecord(d.table_name, d.key, d.value)
		}
	case DB_OPERATE_TYPE_UPDATE:
		{
			this.db.UpdateRecord(d.table_name, d.key, d.value, d.field_list...)
		}
	}
}

func (this *DBOperateManager) CheckAndDo() {
	this.locker.Lock()
	if this.op_list.GetLength() == 0 {
		this.locker.Unlock()
		return
	}
	tmp_list := this.op_list
	this.op_list = &List{}
	this.locker.Unlock()

	node := tmp_list.GetHeadNode()
	for node != nil {
		op_data := node.GetData().(*OpData)
		if op_data != nil {
			if op_data.sql_type == DB_SQL_TYPE_COMMAND {
				if op_data.detail_list != nil && len(op_data.detail_list) > 0 {
					d := op_data.detail_list[0]
					this._op_cmd(d)
				}
			} else {
				if op_data.detail_list != nil && len(op_data.detail_list) > 0 {
					for _, d := range op_data.detail_list {

					}
				}
			}
		}
		node = node.GetNext()
	}
	tmp_list.Clear()
}
