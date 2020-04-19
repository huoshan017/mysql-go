package mysql_manager

import (
	"log"
	"sync"

	mysql_base "github.com/huoshan017/mysql-go/base"
	mysql_generate "github.com/huoshan017/mysql-go/generate"
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
	detail      *mysql_base.OpDetail
	detail_list []*mysql_base.OpDetail
}

type Transaction struct {
	op_mgr      *OperateManager
	detail_list []*mysql_base.OpDetail
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
	this.detail_list = append(this.detail_list, &mysql_base.OpDetail{
		TableName: table_name,
		OpType:    DB_OPERATE_TYPE_INSERT,
		FieldList: field_list,
	})
}

func (this *Transaction) InsertIgnore(table_name string, field_list []*mysql_base.FieldValuePair) {
	this.detail_list = append(this.detail_list, &mysql_base.OpDetail{
		TableName: table_name,
		OpType:    DB_OPERATE_TYPE_INSERT_IGNORE,
		FieldList: field_list,
	})
}

func (this *Transaction) Update(table_name string, key string, value interface{}, field_list []*mysql_base.FieldValuePair) {
	this.detail_list = append(this.detail_list, &mysql_base.OpDetail{
		TableName: table_name,
		OpType:    DB_OPERATE_TYPE_UPDATE,
		Key:       key,
		Value:     value,
		FieldList: field_list,
	})
}

func (this *Transaction) Delete(table_name string, key string, value interface{}) {
	this.detail_list = append(this.detail_list, &mysql_base.OpDetail{
		TableName: table_name,
		OpType:    DB_OPERATE_TYPE_DELETE,
		Key:       key,
		Value:     value,
	})
}

func (this *Transaction) SetDetailList(detail_list []*mysql_base.OpDetail) {
	this.detail_list = detail_list
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
	op_list          *mysql_base.List
	table_op_map     map[string]*table_info
	curr_op_id       uint32
	curr_cmd_op_id   uint32
	curr_trans_op_id uint32
	locker           sync.RWMutex
	db               *mysql_base.Database
	enable           bool
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

func (this *OperateManager) get_table_op_data(table_name string, field_name string, field_value interface{}) *OpData {
	var op_data *OpData
	table_op := this.table_op_map[table_name]
	if table_op != nil && table_op.row_op_map != nil {
		if field_name == table_op.table_primary_field {
			op_data = table_op.row_op_map[field_value]
		}
	}
	return op_data
}

func (this *OperateManager) get_table_op_data_with_field_list(table_name string, field_list []*mysql_base.FieldValuePair) (*OpData, string, interface{}) {
	var value interface{}
	var primary_field string
	table_op := this.table_op_map[table_name]
	if table_op != nil {
		for _, f := range field_list {
			if table_op.table_primary_field == f.Name {
				value = f.Value
				break
			}
		}
		primary_field = table_op.table_primary_field
	}
	var op_data *OpData
	if value != nil && table_op.row_op_map != nil {
		op_data = table_op.row_op_map[value]
	}
	return op_data, primary_field, value
}

func (this *OperateManager) insert_table_op_data(table_name string, field_value interface{}, op_data *OpData) {
	table_op := this.table_op_map[table_name]
	if table_op != nil {
		if table_op.row_op_map == nil {
			table_op.row_op_map = make(map[interface{}]*OpData)
		}
		table_op.row_op_map[field_value] = op_data
	}
}

func (this *OperateManager) reset_table_op_data() {
	if this.table_op_map != nil {
		for _, t := range this.table_op_map {
			t.row_op_map = nil
		}
	}
}

func (this *OperateManager) Insert(table_name string, field_list []*mysql_base.FieldValuePair, ignore bool) {
	this.locker.Lock()
	defer this.locker.Unlock()

	if !this.enable {
		return
	}

	var op_data *OpData
	var field_name string
	var field_value interface{}
	op_data, field_name, field_value = this.get_table_op_data_with_field_list(table_name, field_list)
	if op_data != nil && op_data.detail.OpType != DB_OPERATE_TYPE_DELETE && op_data.id > this.curr_trans_op_id {
		log.Printf("mysql_manager: operate manager insert table %v new row with(field_name:%v, field_value:%v) already exist\n", table_name, field_name, field_value)
		return
	}

	op_data = &OpData{
		id:       this.curr_op_id,
		sql_type: DB_SQL_TYPE_COMMAND,
		detail: &mysql_base.OpDetail{
			TableName: table_name,
			OpType: func() int32 {
				if !ignore {
					return DB_OPERATE_TYPE_INSERT
				} else {
					return DB_OPERATE_TYPE_INSERT_IGNORE
				}
			}(),
			FieldList: field_list,
		},
	}
	this.op_list.Append(op_data)
	this.curr_op_id += 1
	this.insert_table_op_data(table_name, field_value, op_data)
}

func (this *OperateManager) Delete(table_name string, field_name string, field_value interface{}) {
	this.locker.Lock()
	defer this.locker.Unlock()

	if !this.enable {
		return
	}

	op_data := this.get_table_op_data(table_name, field_name, field_value)
	if op_data != nil && op_data.sql_type == DB_SQL_TYPE_COMMAND && op_data.id > this.curr_trans_op_id {
		if op_data.detail != nil {
			// 已经有删除，直接跳过
			if op_data.detail.OpType == DB_OPERATE_TYPE_DELETE {
				return
			}
			// 如果是插入，则直接把命令删除
			if op_data.detail.OpType == DB_OPERATE_TYPE_INSERT {
				this.op_list.Delete(op_data)
				return
			}
			if op_data.detail.OpType == DB_OPERATE_TYPE_UPDATE {
				op_data.detail.OpType = DB_OPERATE_TYPE_DELETE
				if op_data.detail.FieldList != nil {
					op_data.detail.FieldList = nil
				}
				this.op_list.MoveToLast(op_data)
				this.curr_op_id += 1
				return
			}
		}
	}

	op_data = &OpData{
		id:       this.curr_op_id,
		sql_type: DB_SQL_TYPE_COMMAND,
		detail: &mysql_base.OpDetail{
			TableName: table_name,
			OpType:    DB_OPERATE_TYPE_DELETE,
			Key:       field_name,
			Value:     field_value,
		},
	}
	this.op_list.Append(op_data)
	this.curr_op_id += 1
	this.insert_table_op_data(table_name, field_value, op_data)
}

func field_list_cover(field_list1, field_list2 []*mysql_base.FieldValuePair) (merged_list []*mysql_base.FieldValuePair) {
	if field_list1 == nil {
		return
	}

	if field_list2 == nil {
		merged_list = field_list1
		return
	}

	var fm = make(map[string]*mysql_base.FieldValuePair)
	for i := 0; i < len(field_list1); i++ {
		fm[field_list1[i].Name] = field_list1[i]
	}

	merged_list = field_list1
	for _, f2 := range field_list2 {
		if fm[f2.Name] == nil {
			merged_list = append(merged_list, f2)
		}
	}
	return
}

func (this *OperateManager) Update(table_name string, field_name string, field_value interface{}, field_list []*mysql_base.FieldValuePair) {
	this.locker.Lock()
	defer this.locker.Unlock()

	if !this.enable {
		return
	}

	op_data := this.get_table_op_data(table_name, field_name, field_value)
	if op_data != nil && op_data.sql_type == DB_SQL_TYPE_COMMAND && op_data.id > this.curr_trans_op_id {
		if op_data.detail != nil {
			// 已经删除
			if op_data.detail.OpType == DB_OPERATE_TYPE_DELETE {
				return
			}
			if op_data.detail.OpType == DB_OPERATE_TYPE_INSERT || op_data.detail.OpType == DB_OPERATE_TYPE_INSERT_IGNORE || op_data.detail.OpType == DB_OPERATE_TYPE_UPDATE {
				op_data.detail.FieldList = field_list_cover(field_list, op_data.detail.FieldList)
				this.op_list.MoveToLast(op_data)
				this.curr_op_id += 1
			}
			return
		}
	}

	op_data = &OpData{
		id:       this.curr_op_id,
		sql_type: DB_SQL_TYPE_COMMAND,
		detail: &mysql_base.OpDetail{
			TableName: table_name,
			OpType:    DB_OPERATE_TYPE_UPDATE,
			Key:       field_name,
			Value:     field_value,
			FieldList: field_list,
		},
	}
	this.op_list.Append(op_data)
	this.curr_op_id += 1
	this.insert_table_op_data(table_name, field_value, op_data)
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
		id:          this.curr_op_id,
		sql_type:    DB_SQL_TYPE_PROCEDURE,
		detail_list: transaction.detail_list,
	})

	this.curr_trans_op_id = this.curr_op_id
	this.curr_op_id += 1
}

func (this *OperateManager) _op_cmd(d *mysql_base.OpDetail) {
	switch d.OpType {
	case DB_OPERATE_TYPE_INSERT:
		this.db.InsertRecord(d.TableName, d.FieldList...)
	case DB_OPERATE_TYPE_DELETE:
		this.db.DeleteRecord(d.TableName, d.Key, d.Value)
	case DB_OPERATE_TYPE_UPDATE:
		this.db.UpdateRecord(d.TableName, d.Key, d.Value, d.FieldList...)
	case DB_OPERATE_TYPE_INSERT_IGNORE:
		this.db.InsertIgnoreRecord(d.TableName, d.FieldList...)
	}
}

func (this *OperateManager) _op_transaction(dl []*mysql_base.OpDetail) (err error) {
	procedure := this.db.BeginProcedure()
	if procedure == nil {
		return
	}
	for _, d := range dl {
		if d.OpType == DB_OPERATE_TYPE_INSERT {
			err, _ = procedure.InsertRecord(d.TableName, d.FieldList...)
		} else if d.OpType == DB_OPERATE_TYPE_UPDATE {
			err = procedure.UpdateRecord(d.TableName, d.Key, d.Value, d.FieldList...)
		} else if d.OpType == DB_OPERATE_TYPE_DELETE {
			err = procedure.DeleteRecord(d.TableName, d.Key, d.Value)
		} else if d.OpType == DB_OPERATE_TYPE_INSERT_IGNORE {
			err, _ = procedure.InsertIgnoreRecord(d.TableName, d.FieldList...)
		}
		if err != nil {
			procedure.Rollback()
			return
		}
	}
	procedure.Commit()
	return
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
	this.curr_op_id = 0
	this.curr_cmd_op_id = 0
	this.curr_trans_op_id = 0
	return tmp_list
}

func (this *OperateManager) Save() {
	if this._check_op_list_empty() {
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
			if op_data.detail != nil {
				this._op_cmd(op_data.detail)
			}
		} else if op_data.sql_type == DB_SQL_TYPE_PROCEDURE {
			if op_data.detail_list != nil && len(op_data.detail_list) > 0 {
				this._op_transaction(op_data.detail_list)
			}
		}
		node = node.GetNext()
	}
	tmp_list.Clear()
	this.reset_table_op_data()
}
