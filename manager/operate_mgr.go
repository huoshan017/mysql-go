package mysql_manager

import (
	"log"
	"sync"

	"github.com/huoshan017/mysql-go/base"
	"github.com/huoshan017/mysql-go/generate"
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
	field_list []*mysql_base.FieldValuePair
}

type OpData struct {
	id          uint32
	sql_type    int32
	detail      *OpDetail
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

func (this *Transaction) Insert(table_name string, field_list []*mysql_base.FieldValuePair) {
	this.detail_list = append(this.detail_list, &OpDetail{
		table_name: table_name,
		op_type:    DB_OPERATE_TYPE_INSERT,
		field_list: field_list,
	})
}

func (this *Transaction) InsertIgnore(table_name string, field_list []*mysql_base.FieldValuePair) {
	this.detail_list = append(this.detail_list, &OpDetail{
		table_name: table_name,
		op_type:    DB_OPERATE_TYPE_INSERT_IGNORE,
		field_list: field_list,
	})
}

func (this *Transaction) Update(table_name string, key string, value interface{}, field_list []*mysql_base.FieldValuePair) {
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

type table_info struct {
	table_primary_field string
	row_op_map          map[interface{}]*OpData
}

func (this *table_info) init(primary_field string) {
	this.table_primary_field = primary_field
	this.row_op_map = make(map[interface{}]*OpData)
}

type OperateManager struct {
	op_list      *mysql_base.List
	table_op_map map[string]*table_info
	curr_op_id   uint32
	locker       sync.RWMutex
	db           *mysql_base.Database
	enable       bool
}

func (this *OperateManager) Init(db *mysql_base.Database, config_loader *mysql_generate.ConfigLoader) {
	this.op_list = &mysql_base.List{}
	this.db = db
	this.table_op_map = make(map[string]*table_info)
	for _, table := range config_loader.Tables {
		ti := &table_info{}
		ti.init(table.PrimaryKey)
		this.table_op_map[table.Name] = ti
	}
	this.enable = true
}

func (this *OperateManager) GetDB() *mysql_base.Database {
	return this.db
}

func (this *OperateManager) Enable(enable bool) {
	this.locker.Lock()
	defer this.locker.Unlock()
	this.enable = enable
}

func (this *OperateManager) Insert(table_name string, field_list []*mysql_base.FieldValuePair, ignore bool) {
	this.locker.Lock()
	defer this.locker.Unlock()

	if !this.enable {
		return
	}

	this.curr_op_id += 1
	this.op_list.Append(&OpData{
		id:       this.curr_op_id,
		sql_type: DB_SQL_TYPE_COMMAND,
		detail: &OpDetail{
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
	})
}

func (this *OperateManager) Delete(table_name string, field_name string, field_value interface{}) {
	this.locker.Lock()
	defer this.locker.Unlock()

	if !this.enable {
		return
	}

	this.curr_op_id += 1
	this.op_list.Append(&OpData{
		id:       this.curr_op_id,
		sql_type: DB_SQL_TYPE_COMMAND,
		detail: &OpDetail{
			table_name: table_name,
			op_type:    DB_OPERATE_TYPE_DELETE,
			key:        field_name,
			value:      field_value,
		},
	})
}

func (this *OperateManager) Update(table_name string, key string, value interface{}, field_list []*mysql_base.FieldValuePair) {
	this.locker.Lock()
	defer this.locker.Unlock()

	if !this.enable {
		return
	}

	this.curr_op_id += 1
	this.op_list.Append(&OpData{
		id:       this.curr_op_id,
		sql_type: DB_SQL_TYPE_COMMAND,
		detail: &OpDetail{
			table_name: table_name,
			op_type:    DB_OPERATE_TYPE_UPDATE,
			key:        key,
			value:      value,
			field_list: field_list,
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

	this.curr_op_id += 1
	this.op_list.Append(&OpData{
		id:          this.curr_op_id,
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

func (this *OperateManager) _op_transaction(dl []*OpDetail) {
	procedure := this.db.BeginProcedure()
	if procedure == nil {
		return
	}
	for _, d := range dl {
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

func (this *OperateManager) _check_op_list_empty() bool {
	this.locker.RLock()
	defer this.locker.RUnlock()
	return this.op_list.GetLength() == 0
}

func (this *OperateManager) _get_tmp_op_list() *mysql_base.List {
	this.locker.Lock()
	defer this.locker.Unlock()
	if this.op_list.GetLength() == 0 {
		return nil
	}
	tmp_list := this.op_list
	this.op_list = &mysql_base.List{}
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
			d := op_data.detail
			if d != nil {
				this._op_cmd(d)
			}
		} else if op_data.sql_type == DB_SQL_TYPE_PROCEDURE {
			if op_data.detail_list != nil && len(op_data.detail_list) > 0 {
				this._op_transaction(op_data.detail_list)
			}
		}
		node = node.GetNext()
	}
	tmp_list.Clear()
}
