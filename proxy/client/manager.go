package mysql_proxy

import (
	"log"
	"sync/atomic"
	"time"

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

type DB struct {
	read_client   *Client
	write_client  *Client
	db_host_id    int32
	db_host_alias string
	inited        bool
}

func (this *DB) Connect(proxy_address string, db_host_id int32, db_host_alias string) bool {
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
	this.inited = true
	return true
}

func (this *DB) Insert(table_name string, field_pair []*mysql_base.FieldValuePair) {

}

func (this *DB) InsertIgnore(table_name string, field_pair []*mysql_base.FieldValuePair) {
	//this.op_mgr.Insert(table_name, field_pair, true)
}

func (this *DB) Update(table_name string, field_name string, field_value interface{}, field_pair []*mysql_base.FieldValuePair) {
	//this.op_mgr.Update(table_name, field_name, field_value, field_pair)
}

func (this *DB) Delete(table_name string, field_name string, field_value interface{}) {
	//this.op_mgr.Delete(table_name, field_name, field_value)
}

func (this *DB) Select(table_name string, field_name string, field_value interface{}, field_list []string, dest_list []interface{}) bool {
	//return this.database.SelectRecord(table_name, field_name, field_value, field_list, dest_list)
}

func (this *DB) SelectRecords(table_name string, field_name string, field_value interface{}, field_list []string, result_list *mysql_base.QueryResultList) bool {
	//return this.database.SelectRecords(table_name, field_name, field_value, field_list, result_list)
}

func (this *DB) SelectStarRecords(table_name string, field_name string, field_value interface{}, result_list *mysql_base.QueryResultList) bool {
	//return this.database.SelectRecords(table_name, field_name, field_value, nil, result_list)
}

func (this *DB) SelectAllRecords(table_name string, field_list []string, result_list *mysql_base.QueryResultList) bool {
	//return this.database.SelectRecords(table_name, "", nil, field_list, result_list)
}

func (this *DB) SelectFieldNoKey(table_name string, field_name string, result_list *mysql_base.QueryResultList) bool {
	//return this.database.SelectRecords(table_name, "", nil, []string{field_name}, result_list)
}

func (this *DB) SelectRecordsOrderby(table_name string, field_name string, field_value interface{}, order_by string, desc bool, offset, limit int, field_list []string, result_list *mysql_base.QueryResultList) bool {
	//return this.database.SelectRecordsOrderby(table_name, field_name, field_value, order_by, desc, offset, limit, field_list, result_list)
}

/*func (this *DB) NewTransaction() *Transaction {
	return this.op_mgr.NewTransaction()
}*/

func (this *DB) Close() {
	this.database.Close()
}

func (this *DB) EndRun() bool {
	return atomic.CompareAndSwapInt32(&this.state, DB_STATE_RUNNING, DB_STATE_TO_END)
}

func (this *DB) IsEnd() bool {
	return atomic.LoadInt32(&this.state) == DB_STATE_NO_RUN
}

func (this *DB) Save() {
}

func (this *DB) Run() {
}
