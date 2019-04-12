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

type OpData struct {
	table_name string
	op_type    int32
	field_list []*FieldValuePair
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
