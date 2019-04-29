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
	DB_OPERATE_TYPE_SELECT        = iota
	DB_OPERATE_TYPE_INSERT        = 1
	DB_OPERATE_TYPE_DELETE        = 2
	DB_OPERATE_TYPE_UPDATE        = 3
	DB_OPERATE_TYPE_INSERT_IGNORE = 4
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

type Transaction struct {
	op_mgr      *OperateManager
	detail_list []*OpDetail
}

func CreateTransaction(op_mgr *OperateManager) *Transaction {
	return &Transaction{op_mgr: op_mgr}
}

func (this *Transaction) Done() {
	if this.op_mgr != nil {
		this.op_mgr.appendTransaction(this)
	}
}

func (this *Transaction) Insert(table_name string, field_list []*FieldValuePair) {
	this.detail_list = append(this.detail_list, &OpDetail{
		table_name: table_name,
		op_type:    DB_OPERATE_TYPE_INSERT,
		field_list: field_list,
	})
}

func (this *Transaction) InsertIgnore(table_name string, field_list []*FieldValuePair) {
	this.detail_list = append(this.detail_list, &OpDetail{
		table_name: table_name,
		op_type:    DB_OPERATE_TYPE_INSERT_IGNORE,
		field_list: field_list,
	})
}

func (this *Transaction) Update(table_name string, key string, value interface{}, field_list []*FieldValuePair) {
	this.detail_list = append(this.detail_list, &OpDetail{
		table_name: table_name,
		op_type:    DB_OPERATE_TYPE_UPDATE,
		key:        key,
		value:      value,
		field_list: field_list,
	})
}

func (this *Transaction) Delete(table_name string, key string, value interface{}) {
	this.detail_list = append(this.detail_list, &OpDetail{
		table_name: table_name,
		op_type:    DB_OPERATE_TYPE_DELETE,
		key:        key,
		value:      value,
	})
}

type OperateManager struct {
	op_list      *List
	table_op_map map[string]map[interface{}]*OpData
	locker       sync.RWMutex
	db           *Database
	enable       bool
}

func (this *OperateManager) Init(db *Database, tables_name []string) {
	this.op_list = &List{}
	this.db = db
	this.table_op_map = make(map[string]map[interface{}]*OpData)
	for _, tn := range tables_name {
		this.table_op_map[tn] = make(map[interface{}]*OpData)
	}
	this.enable = true
}

func (this *OperateManager) GetDB() *Database {
	return this.db
}

func (this *OperateManager) Enable(enable bool) {
	this.locker.Lock()
	defer this.locker.Unlock()
	this.enable = enable
}

func (this *OperateManager) Insert(table_name string, field_list []*FieldValuePair, ignore bool) {
	this.locker.Lock()
	defer this.locker.Unlock()

	if !this.enable {
		return
	}

	this.op_list.Append(&OpData{
		sql_type: DB_SQL_TYPE_COMMAND,
		detail_list: []*OpDetail{
			&OpDetail{
				table_name: table_name,
				op_type: func() int32 {
					if !ignore {
						return DB_OPERATE_TYPE_INSERT
					} else {
						return DB_OPERATE_TYPE_INSERT_IGNORE
					}
				}(),
				field_list: field_list,
			},
		},
	})
}

func (this *OperateManager) Delete(table_name string, field_name string, field_value interface{}) {
	this.locker.Lock()
	defer this.locker.Unlock()

	if !this.enable {
		return
	}

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

func (this *OperateManager) Update(table_name string, key string, value interface{}, field_list []*FieldValuePair) {
	this.locker.Lock()
	defer this.locker.Unlock()

	if !this.enable {
		return
	}

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

func (this *OperateManager) NewTransaction() *Transaction {
	return CreateTransaction(this)
}

func (this *OperateManager) appendTransaction(transaction *Transaction) {
	this.locker.Lock()
	defer this.locker.Unlock()

	if !this.enable {
		return
	}

	this.op_list.Append(&OpData{
		sql_type:    DB_SQL_TYPE_PROCEDURE,
		detail_list: transaction.detail_list,
	})
}

func (this *OperateManager) _op_cmd(d *OpDetail) {
	switch d.op_type {
	case DB_OPERATE_TYPE_INSERT:
		this.db.InsertRecord(d.table_name, d.field_list...)
	case DB_OPERATE_TYPE_DELETE:
		this.db.DeleteRecord(d.table_name, d.key, d.value)
	case DB_OPERATE_TYPE_UPDATE:
		this.db.UpdateRecord(d.table_name, d.key, d.value, d.field_list...)
	case DB_OPERATE_TYPE_INSERT_IGNORE:
		this.db.InsertIgnoreRecord(d.table_name, d.field_list...)
	}
}

func (this *OperateManager) _check_op_list_empty() bool {
	this.locker.RLock()
	defer this.locker.RUnlock()
	return this.op_list.GetLength() == 0
}

func (this *OperateManager) _get_tmp_op_list() *List {
	this.locker.Lock()
	defer this.locker.Unlock()
	if this.op_list.GetLength() == 0 {
		return nil
	}
	tmp_list := this.op_list
	this.op_list = &List{}
	return tmp_list
}

func (this *OperateManager) Save() {
	if this._check_op_list_empty() {
		log.Printf("DBOperateManager::Save @@@ no operation to execute\n")
		return
	}

	tmp_list := this._get_tmp_op_list()
	if tmp_list == nil {
		return
	}

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
					} else if d.op_type == DB_OPERATE_TYPE_INSERT_IGNORE {
						o, _ = procedure.InsertIgnoreRecord(d.table_name, d.field_list...)
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
