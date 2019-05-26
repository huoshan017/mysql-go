package mysql_proxy

import (
	"log"
	//"sync/atomic"
	"time"

	"github.com/huoshan017/mysql-go/base"
	"github.com/huoshan017/mysql-go/proxy/common"
)

const (
	DEFAULT_CONN_MAX_LIFE_SECONDS = time.Second * 5
	DEFAULT_SAVE_INTERVAL_TIME    = time.Minute * 5
)

const (
	DB_STATE_NO_RUN  = iota
	DB_STATE_RUNNING = 1
	DB_STATE_TO_END  = 2
)

const (
	DB_OPERATE_TYPE_SELECT        = iota
	DB_OPERATE_TYPE_INSERT        = 1
	DB_OPERATE_TYPE_DELETE        = 2
	DB_OPERATE_TYPE_UPDATE        = 3
	DB_OPERATE_TYPE_INSERT_IGNORE = 4
)

type QueryResultList struct {
	rows        [][]interface{}
	cur_row_idx int
}

func CreateQueryResultList(rows [][]interface{}) *QueryResultList {
	return &QueryResultList{
		rows: rows,
	}
}

func (this *QueryResultList) Init(rows [][]interface{}) {
	this.rows = rows
}

func (this *QueryResultList) Close() {
	this.rows = nil
}

func (this *QueryResultList) Get(dest ...interface{}) bool {
	if this.cur_row_idx >= len(this.rows) {
		return false
	}
	row := this.rows[this.cur_row_idx]
	if len(dest) != len(row) {
		log.Printf("mysql-proxy-client: QueryResultList:Get arg dest length must equal to row length\n")
		return false
	}
	for i := 0; i < len(row); i++ {
		if !_copy_reply_value_2_dest(dest[i], row[i]) {
			return false
		}
	}
	this.cur_row_idx += 1
	return true
}

type OpDetail struct {
	table_name string
	op_type    int32
	key        string
	value      interface{}
	field_list []*mysql_base.FieldValuePair
}

type Transaction struct {
	detail_list []*OpDetail
}

func CreateTransaction() *Transaction {
	return &Transaction{}
}

func (this *Transaction) Done() {
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

type DB struct {
	read_client   *Client
	write_client  *Client
	db_host_id    int32
	db_host_alias string
	db_name       string
	inited        bool
}

func (this *DB) Connect(proxy_address string, db_host_id int32, db_host_alias, db_name string) bool {
	client := NewClient()
	if !client.Dial(proxy_address, mysql_proxy_common.CONNECTION_TYPE_ONLY_READ) {
		return false
	}
	this.read_client = client
	client = NewClient()
	if !client.Dial(proxy_address, mysql_proxy_common.CONNECTION_TYPE_WRITE) {
		return false
	}
	this.write_client = client
	this.db_host_id = db_host_id
	this.db_host_alias = db_host_alias
	this.db_name = db_name
	this.inited = true
	return true
}

func (this *DB) _gen_head() *mysql_proxy_common.ArgsHead {
	var head mysql_proxy_common.ArgsHead
	head.SetDBHostId(this.db_host_id)
	head.SetDBHostAlias(this.db_host_alias)
	head.SetDBName(this.db_name)
	return &head
}

func (this *DB) _insert(table_name string, field_pairs []*mysql_base.FieldValuePair, ignore bool) {
	var args = &mysql_proxy_common.InsertRecordArgs{
		Head:            this._gen_head(),
		TableName:       table_name,
		FieldValuePairs: field_pairs,
		Ignore:          ignore,
	}
	var reply mysql_proxy_common.InsertRecordReply
	err := this.write_client.Call("ProxyWriteProc.InsertRecord", args, &reply)
	if err != nil {
		log.Printf("mysql-proxy-client: call InsertRecord err %v\n", err.Error())
	}
}

func (this *DB) Insert(table_name string, field_pairs []*mysql_base.FieldValuePair) {
	this._insert(table_name, field_pairs, true)
}

func (this *DB) InsertIgnore(table_name string, field_pairs []*mysql_base.FieldValuePair) {
	this._insert(table_name, field_pairs, false)
}

func (this *DB) Update(table_name string, field_name string, field_value interface{}, field_pairs []*mysql_base.FieldValuePair) {
	var args = &mysql_proxy_common.UpdateRecordArgs{
		Head:            this._gen_head(),
		TableName:       table_name,
		WhereFieldName:  field_name,
		WhereFieldValue: field_value,
		FieldValuePairs: field_pairs,
	}
	var reply mysql_proxy_common.UpdateRecordReply
	err := this.write_client.Call("ProxyWriteProc.UpdateRecord", args, &reply)
	if err != nil {
		log.Printf("mysql-proxy-client: call UpdateRecord err %v\n", err.Error())
	}
}

func (this *DB) Delete(table_name string, field_name string, field_value interface{}) {
	var args = &mysql_proxy_common.DeleteRecordArgs{
		Head:            this._gen_head(),
		TableName:       table_name,
		WhereFieldName:  field_name,
		WhereFieldValue: field_value,
	}
	var reply mysql_proxy_common.DeleteRecordReply
	err := this.write_client.Call("ProxyWriteProc.DeleteRecord", args, &reply)
	if err != nil {
		log.Printf("mysql-proxy-client: call DeleteRecord err %v\n", err.Error())
	}
}

func (this *DB) Save() {
	var args = &mysql_proxy_common.SaveImmidiateArgs{}
	var reply mysql_proxy_common.SaveImmidiateReply
	err := this.write_client.Call("ProxyWriteProc.Save", args, &reply)
	if err != nil {
		log.Printf("mysql-proxy-client: call Save err %v\n", err.Error())
	}
}

func (this *DB) Select(table_name string, field_name string, field_value interface{}, field_list []string, dest_list []interface{}) bool {
	var args = &mysql_proxy_common.SelectArgs{
		Head:             this._gen_head(),
		TableName:        table_name,
		WhereFieldName:   field_name,
		WhereFieldValue:  field_value,
		SelectFieldNames: field_list,
	}
	var reply mysql_proxy_common.SelectReply
	err := this.read_client.Call("ProxyReadProc.Select", args, &reply)
	if err != nil {
		log.Printf("mysql-proxy-client: call Select err %v\n", err.Error())
		return false
	}
	if len(dest_list) != len(reply.Result) {
		log.Printf("mysql-proxy-client: Select arg dest_list length must equal to SelectReply ResultList length\n")
		return false
	}
	for i := 0; i < len(dest_list); i++ {
		if !_copy_reply_value_2_dest(dest_list[i], reply.Result[i]) {
			return false
		}
	}
	return true
}

func (this *DB) SelectRecords(table_name string, field_name string, field_value interface{}, field_list []string, result_list *QueryResultList) bool {
	var args = &mysql_proxy_common.SelectRecordsArgs{
		Head:             this._gen_head(),
		TableName:        table_name,
		WhereFieldName:   field_name,
		WhereFieldValue:  field_value,
		SelectFieldNames: field_list,
	}
	var reply mysql_proxy_common.SelectRecordsReply
	err := this.read_client.Call("ProxyReadProc.SelectRecords", args, &reply)
	if err != nil {
		log.Printf("mysql-proxy-client: call select records err: %v\n", err.Error())
		return false
	}
	result_list.Init(reply.ResultList)
	return true
}

func (this *DB) SelectAllRecords(table_name string, field_list []string, result_list *QueryResultList) bool {
	var args = &mysql_proxy_common.SelectAllRecordsArgs{
		Head:             this._gen_head(),
		TableName:        table_name,
		SelectFieldNames: field_list,
	}
	var reply mysql_proxy_common.SelectAllRecordsReply
	err := this.read_client.Call("ProxyReadProc.SelectAllRecords", args, &reply)
	if err != nil {
		log.Printf("mysql-proxy-client: call select all records err: %v\n", err.Error())
		return false
	}
	result_list.Init(reply.ResultList)
	return true
}

func (this *DB) SelectField(table_name string, field_name string, dest_list []interface{}) bool {
	var args = &mysql_proxy_common.SelectFieldArgs{
		Head:            this._gen_head(),
		TableName:       table_name,
		SelectFieldName: field_name,
	}
	var reply mysql_proxy_common.SelectFieldReply
	err := this.read_client.Call("ProxyReadProc.SelectField", args, &reply)
	if err != nil {
		log.Printf("mysql-proxy-client: call select field err: %v\n", err.Error())
		return false
	}
	if len(dest_list) != len(reply.ResultList) {
		log.Printf("mysql-proxy-client: SelectFieldNokey arg dest_list length must equal to SelectReply ResultList length\n")
		return false
	}
	for i := 0; i < len(dest_list); i++ {
		if !_copy_reply_value_2_dest(dest_list[i], reply.ResultList[i]) {
			return false
		}
	}
	return true
}

func (this *DB) SelectRecordsOrderby(table_name string, field_name string, field_value interface{}, order_by string, desc bool, offset, limit int, field_list []string, result_list *QueryResultList) bool {
	var args = &mysql_proxy_common.SelectRecordsOrderbyArgs{
		Head:             this._gen_head(),
		TableName:        table_name,
		WhereFieldName:   field_name,
		WhereFieldValue:  field_value,
		SelectFieldNames: field_list,
		Orderby:          order_by,
		Desc:             desc,
		Offset:           offset,
		Limit:            limit,
	}
	var reply mysql_proxy_common.SelectRecordsOrderbyReply
	err := this.read_client.Call("ProxyReadProc.SelectRecordsOrderby", args, &reply)
	if err != nil {
		log.Printf("mysql-proxy-client: call select records order by err: %v\n", err.Error())
		return false
	}
	result_list.Init(reply.ResultList)
	return true
}

func (this *DB) NewTransaction() *Transaction {
	return CreateTransaction()
}

func (this *DB) Close() {
	this.read_client.Close()
	this.write_client.Close()
}

func (this *DB) Run() {
	this.read_client.Run()
	this.write_client.Run()
}
