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
	DBHostId int32
	DBName   string
}

func (this *ArgsHead) SetDBHostId(db_host_id int32) {
	this.DBHostId = db_host_id
}

func (this *ArgsHead) SetDBName(db_name string) {
	this.DBName = db_name
}

func (this *ArgsHead) GetDBHostId() int32 {
	return this.DBHostId
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

// select records count
type SelectRecordsCountArgs struct {
	Head            *ArgsHead
	TableName       string
	WhereFieldName  string
	WhereFieldValue interface{}
}

type SelectRecordsCountReply struct {
	Count int32
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

type SelectRecordsMapReply struct {
	ResultMap map[interface{}][]interface{}
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

type SelectAllRecordsMapReply struct {
	ResultMap map[interface{}][]interface{}
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

type SelectRecordsMapConditionReply struct {
	ResultMap map[interface{}][]interface{}
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

// end
type EndArgs struct {
}

type EndReply struct {
}
