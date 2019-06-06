package mysql_proxy_common

import (
	"github.com/huoshan017/mysql-go/base"
)

const (
	CONNECTION_TYPE_NONE      = iota
	CONNECTION_TYPE_ONLY_READ = 1
	CONNECTION_TYPE_WRITE     = 2
)

type PingArgs struct {
}

type PongReply struct {
}

type ArgsHead struct {
	DBHostId    int32
	DBHostAlias string
	DBName      string
}

func (this *ArgsHead) SetDBHostId(db_host_id int32) {
	this.DBHostId = db_host_id
}

func (this *ArgsHead) SetDBHostAlias(db_host_alias string) {
	this.DBHostAlias = db_host_alias
}

func (this *ArgsHead) SetDBName(db_name string) {
	this.DBName = db_name
}

func (this *ArgsHead) GetDBHostId() int32 {
	return this.DBHostId
}

func (this *ArgsHead) GetDBHostAlias() string {
	return this.DBHostAlias
}

func (this *ArgsHead) GetDBName() string {
	return this.DBName
}

// select
type SelectArgs struct {
	Head             *ArgsHead
	TableName        string
	WhereFieldName   string
	WhereFieldValue  interface{}
	SelectFieldNames []string
}

type SelectReply struct {
	Result []interface{}
}

// select records
type SelectRecordsArgs struct {
	Head             *ArgsHead
	TableName        string
	WhereFieldName   string
	WhereFieldValue  interface{}
	SelectFieldNames []string
}

type SelectRecordsReply struct {
	ResultList [][]interface{}
}

// select all records
type SelectAllRecordsArgs struct {
	Head             *ArgsHead
	TableName        string
	SelectFieldNames []string
}

type SelectAllRecordsReply struct {
	ResultList [][]interface{}
}

// select field
type SelectFieldArgs struct {
	Head            *ArgsHead
	TableName       string
	SelectFieldName string
}

type SelectFieldReply struct {
	ResultList []interface{}
}

// select field map
type SelectFieldMapArgs struct {
	Head            *ArgsHead
	TableName       string
	SelectFieldName string
}

type SelectFieldMapReply struct {
	ResultMap map[interface{}]bool
}

// select records order by
type SelectRecordsConditionArgs struct {
	Head             *ArgsHead
	TableName        string
	WhereFieldName   string
	WhereFieldValue  interface{}
	SelectFieldNames []string
	SelCond          *mysql_base.SelectCondition
}

type SelectRecordsConditionReply struct {
	ResultList [][]interface{}
}

// insert record
type InsertRecordArgs struct {
	Head            *ArgsHead
	TableName       string
	FieldValuePairs []*mysql_base.FieldValuePair
	Ignore          bool
}

type InsertRecordReply struct {
}

// update record
type UpdateRecordArgs struct {
	Head            *ArgsHead
	TableName       string
	WhereFieldName  string
	WhereFieldValue interface{}
	FieldValuePairs []*mysql_base.FieldValuePair
}

type UpdateRecordReply struct {
}

// delete record
type DeleteRecordArgs struct {
	Head            *ArgsHead
	TableName       string
	WhereFieldName  string
	WhereFieldValue interface{}
}

type DeleteRecordReply struct {
}

// save immidiate
type SaveImmidiateArgs struct {
	Head *ArgsHead
}

type SaveImmidiateReply struct {
}

// transaction
type CommitTransactionArgs struct {
	Head    *ArgsHead
	Details []*mysql_base.OpDetail
}

type CommitTransactionReply struct {
}
