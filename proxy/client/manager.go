package mysql_proxy

import (
	"log"
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
		if !mysql_base.CopySrcValue2Dest(dest[i], row[i]) {
			return false
		}
	}
	this.cur_row_idx += 1
	return true
}

type Transaction struct {
	db          *DB
	detail_list []*mysql_base.OpDetail
}

func CreateTransaction(db *DB) *Transaction {
	return &Transaction{db: db}
}

func (this *Transaction) Done() {
	var args = &mysql_proxy_common.CommitTransactionArgs{
		Head:    this.db._gen_head(),
		Details: this.detail_list,
	}
	var reply mysql_proxy_common.CommitTransactionReply
	err := this.db.write_client.Call("ProxyWriteProc.CommitTransaction", args, &reply)
	if err != nil {
		log.Printf("mysql-proxy-client: call CommitTransaction err %v\n", err.Error())
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

type DB struct {
	read_client   *client
	write_client  *client
	db_host_id    int32
	db_host_alias string
	db_name       string
	inited        bool
}

func (this *DB) Connect(proxy_address string, db_host_id int32, db_host_alias, db_name string) error {
	client := new_client()
	err := client.Dial(proxy_address, mysql_proxy_common.CONNECTION_TYPE_ONLY_READ)
	if err != nil {
		return err
	}
	this.read_client = client
	client = new_client()
	err = client.Dial(proxy_address, mysql_proxy_common.CONNECTION_TYPE_WRITE)
	if err != nil {
		return err
	}
	this.write_client = client
	this.db_host_id = db_host_id
	this.db_host_alias = db_host_alias
	this.db_name = db_name
	this.inited = true
	return nil
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

func (this *DB) End() {
	var args mysql_proxy_common.EndArgs
	var reply mysql_proxy_common.EndReply
	err := this.write_client.Call("ProxyWriteProc.End", &args, &reply)
	if err != nil {
		log.Printf("mysql-proxy-client: call End err %v\n", err.Error())
	}
}

func (this *DB) Select(table_name string, field_name string, field_value interface{}, field_list []string, dest_list []interface{}) error {
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
		return err
	}
	if len(dest_list) != len(reply.Result) {
		log.Printf("mysql-proxy-client: Select arg dest_list length must equal to SelectReply ResultList length\n")
		return mysql_base.ErrArgumentInvalid
	}
	for i := 0; i < len(dest_list); i++ {
		if !mysql_base.CopySrcValue2Dest(dest_list[i], reply.Result[i]) {
			return mysql_base.ErrInternal
		}
	}
	return nil
}

func (this *DB) SelectRecordsCount(table_name, field_name string, field_value interface{}) (count int32, err error) {
	var args = mysql_proxy_common.SelectRecordsCountArgs{
		Head:            this._gen_head(),
		TableName:       table_name,
		WhereFieldName:  field_name,
		WhereFieldValue: field_value,
	}
	var reply mysql_proxy_common.SelectRecordsCountReply
	err = this.read_client.Call("ProxyReadProc.SelectRecordsCount", &args, &reply)
	if err != nil {
		return
	}
	count = reply.Count
	return
}

func (this *DB) SelectRecords(table_name string, field_name string, field_value interface{}, field_list []string, result_list *QueryResultList) error {
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
		return err
	}
	result_list.Init(reply.ResultList)
	return nil
}

func (this *DB) SelectRecordsMap(table_name string, field_name string, field_value interface{}, field_list []string) (records_map map[interface{}][]interface{}, err error) {
	var args = mysql_proxy_common.SelectRecordsArgs{
		Head:             this._gen_head(),
		TableName:        table_name,
		WhereFieldName:   field_name,
		WhereFieldValue:  field_value,
		SelectFieldNames: field_list,
	}
	var reply mysql_proxy_common.SelectRecordsMapReply
	err = this.read_client.Call("ProxyReadProc.SelectRecordsMap", &args, &reply)
	if err != nil {
		log.Printf("mysql-proxy-client: call select records map err: %v\n", err.Error())
		return nil, err
	}
	return reply.ResultMap, nil
}

func (this *DB) SelectAllRecords(table_name string, field_list []string, result_list *QueryResultList) error {
	var args = &mysql_proxy_common.SelectAllRecordsArgs{
		Head:             this._gen_head(),
		TableName:        table_name,
		SelectFieldNames: field_list,
	}
	var reply mysql_proxy_common.SelectAllRecordsReply
	err := this.read_client.Call("ProxyReadProc.SelectAllRecords", args, &reply)
	if err != nil {
		log.Printf("mysql-proxy-client: call select all records err: %v\n", err.Error())
		return err
	}
	result_list.Init(reply.ResultList)
	return nil
}

func (this *DB) SelectAllRecordsMap(table_name string, field_list []string) (records_map map[interface{}][]interface{}, err error) {
	var args = mysql_proxy_common.SelectAllRecordsArgs{
		Head:             this._gen_head(),
		TableName:        table_name,
		SelectFieldNames: field_list,
	}
	var reply mysql_proxy_common.SelectAllRecordsMapReply
	err = this.read_client.Call("ProxyReadProc.SelectAllRecordsMap", &args, &reply)
	if err != nil {
		log.Printf("mysql-proxy-client: call select all records map err: %v\n", err.Error())
		return nil, err
	}
	return reply.ResultMap, nil
}

func (this *DB) SelectField(table_name string, field_name string) ([]interface{}, error) {
	var args = &mysql_proxy_common.SelectFieldArgs{
		Head:            this._gen_head(),
		TableName:       table_name,
		SelectFieldName: field_name,
	}
	var reply mysql_proxy_common.SelectFieldReply
	err := this.read_client.Call("ProxyReadProc.SelectField", args, &reply)
	if err != nil {
		log.Printf("mysql-proxy-client: call select field err: %v\n", err.Error())
		return nil, err
	}
	return reply.ResultList, nil
}

func (this *DB) SelectFieldMap(table_name string, field_name string) (map[interface{}]bool, error) {
	var args = mysql_proxy_common.SelectFieldArgs{
		Head:            this._gen_head(),
		TableName:       table_name,
		SelectFieldName: field_name,
	}
	var reply mysql_proxy_common.SelectFieldMapReply
	err := this.read_client.Call("ProxyReadProc.SelectFieldMap", &args, &reply)
	if err != nil {
		log.Printf("mysql-proxy-client: call select field map err: %v\n", err.Error())
		return nil, err
	}
	return reply.ResultMap, nil
}

func (this *DB) SelectRecordsCondition(table_name string, field_name string, field_value interface{}, sel_cond *mysql_base.SelectCondition, field_list []string, result_list *QueryResultList) error {
	var args = &mysql_proxy_common.SelectRecordsConditionArgs{
		Head:             this._gen_head(),
		TableName:        table_name,
		WhereFieldName:   field_name,
		WhereFieldValue:  field_value,
		SelectFieldNames: field_list,
		SelCond:          sel_cond,
	}
	var reply mysql_proxy_common.SelectRecordsConditionReply
	err := this.read_client.Call("ProxyReadProc.SelectRecordsCondition", args, &reply)
	if err != nil {
		log.Printf("mysql-proxy-client: call select records condition err: %v\n", err.Error())
		return err
	}
	result_list.Init(reply.ResultList)
	return nil
}

func (this *DB) SelectRecordsMapCondition(table_name, field_name string, field_value interface{}, sel_cond *mysql_base.SelectCondition, field_list []string) (records_map map[interface{}][]interface{}, err error) {
	var args = mysql_proxy_common.SelectRecordsConditionArgs{
		Head:             this._gen_head(),
		TableName:        table_name,
		WhereFieldName:   field_name,
		WhereFieldValue:  field_value,
		SelectFieldNames: field_list,
		SelCond:          sel_cond,
	}
	var reply mysql_proxy_common.SelectRecordsMapConditionReply
	err = this.read_client.Call("ProxyReadProc.SelectRecordsMapCondition", &args, &reply)
	if err != nil {
		log.Printf("mysql-proxy-client: call select records map condition err: %v\n", err.Error())
		return nil, err
	}
	return reply.ResultMap, nil
}

func (this *DB) NewTransaction() *Transaction {
	return CreateTransaction(this)
}

func (this *DB) Close() {
	this.read_client.Close()
	this.write_client.Close()
}

func (this *DB) GoRun() {
	this.read_client.GoRun()
	this.write_client.GoRun()
}
