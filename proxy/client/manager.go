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

func (p *QueryResultList) Init(rows [][]interface{}) {
	p.rows = rows
}

func (p *QueryResultList) Close() {
	p.rows = nil
}

func (p *QueryResultList) Get(dest ...interface{}) bool {
	if p.cur_row_idx >= len(p.rows) {
		return false
	}
	row := p.rows[p.cur_row_idx]
	if len(dest) != len(row) {
		log.Printf("mysql-proxy-client: QueryResultList:Get arg dest length must equal to row length\n")
		return false
	}
	for i := 0; i < len(row); i++ {
		if !mysql_base.CopySrcValue2Dest(dest[i], row[i]) {
			return false
		}
	}
	p.cur_row_idx += 1
	return true
}

type Transaction struct {
	proxy       *DB
	host_id     int32
	db_name     string
	detail_list []*mysql_base.OpDetail
}

func CreateTransaction(host_id int32, db_name string, proxy *DB) *Transaction {
	return &Transaction{
		proxy:   proxy,
		host_id: host_id,
		db_name: db_name,
	}
}

func (p *Transaction) Done() {
	var args = &mysql_proxy_common.CommitTransactionArgs{
		Head: &mysql_proxy_common.ArgsHead{
			DBHostId: p.host_id,
			DBName:   p.db_name,
		},
		Details: p.detail_list,
	}
	var reply mysql_proxy_common.CommitTransactionReply
	err := p.proxy.write_client.Call("ProxyWriteProc.CommitTransaction", args, &reply)
	if err != nil {
		log.Printf("mysql-proxy-client: call CommitTransaction err %v\n", err.Error())
	}
}

func (p *Transaction) Insert(table_name string, field_list []*mysql_base.FieldValuePair) {
	p.detail_list = append(p.detail_list, &mysql_base.OpDetail{
		TableName: table_name,
		OpType:    DB_OPERATE_TYPE_INSERT,
		FieldList: field_list,
	})
}

func (p *Transaction) InsertIgnore(table_name string, field_list []*mysql_base.FieldValuePair) {
	p.detail_list = append(p.detail_list, &mysql_base.OpDetail{
		TableName: table_name,
		OpType:    DB_OPERATE_TYPE_INSERT_IGNORE,
		FieldList: field_list,
	})
}

func (p *Transaction) Update(table_name string, key string, value interface{}, field_list []*mysql_base.FieldValuePair) {
	p.detail_list = append(p.detail_list, &mysql_base.OpDetail{
		TableName: table_name,
		OpType:    DB_OPERATE_TYPE_UPDATE,
		Key:       key,
		Value:     value,
		FieldList: field_list,
	})
}

func (p *Transaction) Delete(table_name string, key string, value interface{}) {
	p.detail_list = append(p.detail_list, &mysql_base.OpDetail{
		TableName: table_name,
		OpType:    DB_OPERATE_TYPE_DELETE,
		Key:       key,
		Value:     value,
	})
}

type DB struct {
	read_client  *client
	write_client *client
}

func (p *DB) Connect(proxy_address string) error {
	client := new_client()
	err := client.Dial(proxy_address, mysql_proxy_common.CONNECTION_TYPE_ONLY_READ)
	if err != nil {
		return err
	}
	p.read_client = client
	client = new_client()
	err = client.Dial(proxy_address, mysql_proxy_common.CONNECTION_TYPE_WRITE)
	if err != nil {
		return err
	}
	p.write_client = client
	return nil
}

func (p *DB) _insert(host_id int32, db_name, table_name string, field_pairs []*mysql_base.FieldValuePair, ignore bool) {
	var args = &mysql_proxy_common.InsertRecordArgs{
		Head: &mysql_proxy_common.ArgsHead{
			DBHostId: host_id,
			DBName:   db_name,
		},
		TableName:       table_name,
		FieldValuePairs: field_pairs,
		Ignore:          ignore,
	}
	var reply mysql_proxy_common.InsertRecordReply
	err := p.write_client.Call("ProxyWriteProc.InsertRecord", args, &reply)
	if err != nil {
		log.Printf("mysql-proxy-client: call InsertRecord err %v\n", err.Error())
	}
}

func (p *DB) Insert(host_id int32, db_name, table_name string, field_pairs []*mysql_base.FieldValuePair) {
	p._insert(host_id, db_name, table_name, field_pairs, true)
}

func (p *DB) InsertIgnore(host_id int32, db_name, table_name string, field_pairs []*mysql_base.FieldValuePair) {
	p._insert(host_id, db_name, table_name, field_pairs, false)
}

func (p *DB) Update(host_id int32, db_name, table_name string, field_name string, field_value interface{}, field_pairs []*mysql_base.FieldValuePair) {
	var args = &mysql_proxy_common.UpdateRecordArgs{
		Head: &mysql_proxy_common.ArgsHead{
			DBHostId: host_id,
			DBName:   db_name,
		},
		TableName:       table_name,
		WhereFieldName:  field_name,
		WhereFieldValue: field_value,
		FieldValuePairs: field_pairs,
	}
	var reply mysql_proxy_common.UpdateRecordReply
	err := p.write_client.Call("ProxyWriteProc.UpdateRecord", args, &reply)
	if err != nil {
		log.Printf("mysql-proxy-client: call UpdateRecord err %v\n", err.Error())
	}
}

func (p *DB) Delete(host_id int32, db_name, table_name string, field_name string, field_value interface{}) {
	var args = &mysql_proxy_common.DeleteRecordArgs{
		Head: &mysql_proxy_common.ArgsHead{
			DBHostId: host_id,
			DBName:   db_name,
		},
		TableName:       table_name,
		WhereFieldName:  field_name,
		WhereFieldValue: field_value,
	}
	var reply mysql_proxy_common.DeleteRecordReply
	err := p.write_client.Call("ProxyWriteProc.DeleteRecord", args, &reply)
	if err != nil {
		log.Printf("mysql-proxy-client: call DeleteRecord err %v\n", err.Error())
	}
}

func (p *DB) Save() {
	var args = &mysql_proxy_common.SaveImmidiateArgs{}
	var reply mysql_proxy_common.SaveImmidiateReply
	err := p.write_client.Call("ProxyWriteProc.Save", args, &reply)
	if err != nil {
		log.Printf("mysql-proxy-client: call Save err %v\n", err.Error())
	}
}

func (p *DB) End() {
	var args mysql_proxy_common.EndArgs
	var reply mysql_proxy_common.EndReply
	err := p.write_client.Call("ProxyWriteProc.End", &args, &reply)
	if err != nil {
		log.Printf("mysql-proxy-client: call End err %v\n", err.Error())
	}
}

func (p *DB) Select(host_id int32, db_name, table_name string, field_name string, field_value interface{}, field_list []string, dest_list []interface{}) error {
	var args = &mysql_proxy_common.SelectArgs{
		Head: &mysql_proxy_common.ArgsHead{
			DBHostId: host_id,
			DBName:   db_name,
		},
		TableName:        table_name,
		WhereFieldName:   field_name,
		WhereFieldValue:  field_value,
		SelectFieldNames: field_list,
	}
	var reply mysql_proxy_common.SelectReply
	err := p.read_client.Call("ProxyReadProc.Select", args, &reply)
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

func (p *DB) SelectRecordsCount(host_id int32, db_name, table_name, field_name string, field_value interface{}) (count int32, err error) {
	var args = mysql_proxy_common.SelectRecordsCountArgs{
		Head: &mysql_proxy_common.ArgsHead{
			DBHostId: host_id,
			DBName:   db_name,
		},
		TableName:       table_name,
		WhereFieldName:  field_name,
		WhereFieldValue: field_value,
	}
	var reply mysql_proxy_common.SelectRecordsCountReply
	err = p.read_client.Call("ProxyReadProc.SelectRecordsCount", &args, &reply)
	if err != nil {
		return
	}
	count = reply.Count
	return
}

func (p *DB) SelectRecords(host_id int32, db_name, table_name string, field_name string, field_value interface{}, field_list []string, result_list *QueryResultList) error {
	var args = &mysql_proxy_common.SelectRecordsArgs{
		Head: &mysql_proxy_common.ArgsHead{
			DBHostId: host_id,
			DBName:   db_name,
		},
		TableName:        table_name,
		WhereFieldName:   field_name,
		WhereFieldValue:  field_value,
		SelectFieldNames: field_list,
	}
	var reply mysql_proxy_common.SelectRecordsReply
	err := p.read_client.Call("ProxyReadProc.SelectRecords", args, &reply)
	if err != nil {
		log.Printf("mysql-proxy-client: call select records err: %v\n", err.Error())
		return err
	}
	result_list.Init(reply.ResultList)
	return nil
}

func (p *DB) SelectRecordsMap(host_id int32, db_name, table_name string, field_name string, field_value interface{}, field_list []string) (records_map map[interface{}][]interface{}, err error) {
	var args = mysql_proxy_common.SelectRecordsArgs{
		Head: &mysql_proxy_common.ArgsHead{
			DBHostId: host_id,
			DBName:   db_name,
		},
		TableName:        table_name,
		WhereFieldName:   field_name,
		WhereFieldValue:  field_value,
		SelectFieldNames: field_list,
	}
	var reply mysql_proxy_common.SelectRecordsMapReply
	err = p.read_client.Call("ProxyReadProc.SelectRecordsMap", &args, &reply)
	if err != nil {
		log.Printf("mysql-proxy-client: call select records map err: %v\n", err.Error())
		return nil, err
	}
	return reply.ResultMap, nil
}

func (p *DB) SelectAllRecords(host_id int32, db_name, table_name string, field_list []string, result_list *QueryResultList) error {
	var args = &mysql_proxy_common.SelectAllRecordsArgs{
		Head: &mysql_proxy_common.ArgsHead{
			DBHostId: host_id,
			DBName:   db_name,
		},
		TableName:        table_name,
		SelectFieldNames: field_list,
	}
	var reply mysql_proxy_common.SelectAllRecordsReply
	err := p.read_client.Call("ProxyReadProc.SelectAllRecords", args, &reply)
	if err != nil {
		log.Printf("mysql-proxy-client: call select all records err: %v\n", err.Error())
		return err
	}
	result_list.Init(reply.ResultList)
	return nil
}

func (p *DB) SelectAllRecordsMap(host_id int32, db_name, table_name string, field_list []string) (records_map map[interface{}][]interface{}, err error) {
	var args = mysql_proxy_common.SelectAllRecordsArgs{
		Head: &mysql_proxy_common.ArgsHead{
			DBHostId: host_id,
			DBName:   db_name,
		},
		TableName:        table_name,
		SelectFieldNames: field_list,
	}
	var reply mysql_proxy_common.SelectAllRecordsMapReply
	err = p.read_client.Call("ProxyReadProc.SelectAllRecordsMap", &args, &reply)
	if err != nil {
		log.Printf("mysql-proxy-client: call select all records map err: %v\n", err.Error())
		return nil, err
	}
	return reply.ResultMap, nil
}

func (p *DB) SelectField(host_id int32, db_name, table_name string, field_name string) ([]interface{}, error) {
	var args = &mysql_proxy_common.SelectFieldArgs{
		Head: &mysql_proxy_common.ArgsHead{
			DBHostId: host_id,
			DBName:   db_name,
		},
		TableName:       table_name,
		SelectFieldName: field_name,
	}
	var reply mysql_proxy_common.SelectFieldReply
	err := p.read_client.Call("ProxyReadProc.SelectField", args, &reply)
	if err != nil {
		log.Printf("mysql-proxy-client: call select field err: %v\n", err.Error())
		return nil, err
	}
	return reply.ResultList, nil
}

func (p *DB) SelectFieldMap(host_id int32, db_name, table_name string, field_name string) (map[interface{}]bool, error) {
	var args = mysql_proxy_common.SelectFieldArgs{
		Head: &mysql_proxy_common.ArgsHead{
			DBHostId: host_id,
			DBName:   db_name,
		},
		TableName:       table_name,
		SelectFieldName: field_name,
	}
	var reply mysql_proxy_common.SelectFieldMapReply
	err := p.read_client.Call("ProxyReadProc.SelectFieldMap", &args, &reply)
	if err != nil {
		log.Printf("mysql-proxy-client: call select field map err: %v\n", err.Error())
		return nil, err
	}
	return reply.ResultMap, nil
}

func (p *DB) SelectRecordsCondition(host_id int32, db_name, table_name string, field_name string, field_value interface{}, sel_cond *mysql_base.SelectCondition, field_list []string, result_list *QueryResultList) error {
	var args = &mysql_proxy_common.SelectRecordsConditionArgs{
		Head: &mysql_proxy_common.ArgsHead{
			DBHostId: host_id,
			DBName:   db_name,
		},
		TableName:        table_name,
		WhereFieldName:   field_name,
		WhereFieldValue:  field_value,
		SelectFieldNames: field_list,
		SelCond:          sel_cond,
	}
	var reply mysql_proxy_common.SelectRecordsConditionReply
	err := p.read_client.Call("ProxyReadProc.SelectRecordsCondition", args, &reply)
	if err != nil {
		log.Printf("mysql-proxy-client: call select records condition err: %v\n", err.Error())
		return err
	}
	result_list.Init(reply.ResultList)
	return nil
}

func (p *DB) SelectRecordsMapCondition(host_id int32, db_name, table_name, field_name string, field_value interface{}, sel_cond *mysql_base.SelectCondition, field_list []string) (records_map map[interface{}][]interface{}, err error) {
	var args = mysql_proxy_common.SelectRecordsConditionArgs{
		Head: &mysql_proxy_common.ArgsHead{
			DBHostId: host_id,
			DBName:   db_name,
		},
		TableName:        table_name,
		WhereFieldName:   field_name,
		WhereFieldValue:  field_value,
		SelectFieldNames: field_list,
		SelCond:          sel_cond,
	}
	var reply mysql_proxy_common.SelectRecordsMapConditionReply
	err = p.read_client.Call("ProxyReadProc.SelectRecordsMapCondition", &args, &reply)
	if err != nil {
		log.Printf("mysql-proxy-client: call select records map condition err: %v\n", err.Error())
		return nil, err
	}
	return reply.ResultMap, nil
}

func (p *DB) NewTransaction(host_id int32, db_name string) *Transaction {
	return CreateTransaction(host_id, db_name, p)
}

func (p *DB) Close() {
	p.read_client.Close()
	p.write_client.Close()
}

func (p *DB) RunBackground() {
	p.read_client.RunBackground()
	p.write_client.RunBackground()
}

func (p *DB) Run() {
	p.read_client.Run()
	p.write_client.Run()
}
