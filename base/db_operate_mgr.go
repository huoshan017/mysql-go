package mysql_base

import (
	"log"
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

type ProcedureOpList struct {
	detail_list []*OpDetail
}

func (this *ProcedureOpList) Insert(table_name string, field_list []*FieldValuePair) {
	this.detail_list = append(this.detail_list, &OpDetail{
		table_name: table_name,
		op_type:    DB_OPERATE_TYPE_INSERT,
		field_list: field_list,
	})
}

func (this *ProcedureOpList) Update(table_name string, key string, value interface{}, field_list []*FieldValuePair) {
	this.detail_list = append(this.detail_list, &OpDetail{
		table_name: table_name,
		op_type:    DB_OPERATE_TYPE_UPDATE,
		key:        key,
		value:      value,
		field_list: field_list,
	})
}

func (this *ProcedureOpList) Delete(table_name string, key string, value interface{}) {
	this.detail_list = append(this.detail_list, &OpDetail{
		table_name: table_name,
		op_type:    DB_OPERATE_TYPE_DELETE,
		key:        key,
		value:      value,
	})
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

func (this *DBOperateManager) GetDB() *Database {
	return this.db
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

func (this *DBOperateManager) AppendProcedure(procedure *ProcedureOpList) {
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

// 在一个goroutine中执行
func (this *DBOperateManager) Save() {
	this.locker.RLock()
	if this.op_list.GetLength() == 0 {
		log.Printf("DBOperateManager::Save @@@ not operation to execute\n")
		this.locker.RUnlock()
		return
	}
	this.locker.RUnlock()

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
		if op_data == nil {
			node = node.GetNext()
		}

		if op_data.sql_type == DB_SQL_TYPE_COMMAND {
			if op_data.detail_list != nil && len(op_data.detail_list) > 0 {
				d := op_data.detail_list[0]
				this._op_cmd(d)
			}
		} else if op_data.sql_type == DB_SQL_TYPE_PROCEDURE {
			if op_data.detail_list != nil && len(op_data.detail_list) > 0 {
				procedure := this.db.BeginProcedure()
				if procedure == nil {
					continue
				}
				for _, d := range op_data.detail_list {
					var o bool
					if d.op_type == DB_OPERATE_TYPE_INSERT {
						o, _ = procedure.InsertRecord(d.table_name, d.field_list...)
					} else if d.op_type == DB_OPERATE_TYPE_UPDATE {
						o = procedure.UpdateRecord(d.table_name, d.key, d.value, d.field_list...)
					} else if d.op_type == DB_OPERATE_TYPE_DELETE {
						o = procedure.DeleteRecord(d.table_name, d.key, d.value)
					}
					if !o {
						procedure.Rollback()
						break
					}
				}
				procedure.Commit()
			}
		}
		node = node.GetNext()
	}
	tmp_list.Clear()
}
