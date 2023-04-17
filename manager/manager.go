package mysql_manager

import (
	"sync/atomic"
	"time"

	mysql_base "github.com/huoshan017/mysql-go/base"
	mysql_generate "github.com/huoshan017/mysql-go/generate"
	"github.com/huoshan017/mysql-go/log"
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
	config_loader *mysql_generate.ConfigLoader
	database      mysql_base.Database
	op_mgr        OperateManager
	save_interval time.Duration
	state         int32
}

/*func (db *DB) LoadConfig(config_path string) bool {
	config_loader := &mysql_generate.ConfigLoader{}
	if !config_loader.Load(config_path) {
		log.Printf("load config %v failed\n", config_path)
		return false
	}
	db.config_loader = config_loader
	return true
}*/

func (db *DB) AttachConfig(config_loader *mysql_generate.ConfigLoader) {
	db.config_loader = config_loader
}

func (db *DB) GetConfigLoader() *mysql_generate.ConfigLoader {
	return db.config_loader
}

func (db *DB) SetConfigLoader(configLoader *mysql_generate.ConfigLoader) error {
	db.config_loader = configLoader
	for _, t := range db.config_loader.Tables {
		err := db.database.LoadTable(t)
		if err != nil {
			log.Infof("load table %v config failed\n", t.Name)
			return err
		}
	}
	db.op_mgr.Init(&db.database, db.config_loader)
	return nil
}

func (db *DB) Connect(dbhost, dbuser, dbpassword, dbname string) error {
	err := db.database.Open(dbhost, dbuser, dbpassword, dbname)
	if err != nil {
		log.Infof("open database err %v\n", err.Error())
		return err
	}
	db.database.SetMaxLifeTime(DEFAULT_CONN_MAX_LIFE_SECONDS)
	db.save_interval = DEFAULT_SAVE_INTERVAL_TIME
	return nil
}

func (db *DB) SetConnectLifeTime(d time.Duration) {
	db.database.SetMaxLifeTime(d)
}

func (db *DB) SetSaveIntervalTime(d time.Duration) {
	db.save_interval = d
}

func (db *DB) Insert(table_name string, field_pair []*mysql_base.FieldValuePair) {
	db.op_mgr.Insert(table_name, field_pair, false)
}

func (db *DB) InsertIgnore(table_name string, field_pair []*mysql_base.FieldValuePair) {
	db.op_mgr.Insert(table_name, field_pair, true)
}

func (db *DB) Update(table_name string, field_name string, field_value interface{}, field_pair []*mysql_base.FieldValuePair) {
	db.op_mgr.Update(table_name, field_name, field_value, field_pair)
}

func (db *DB) Delete(table_name string, field_name string, field_value interface{}) {
	db.op_mgr.Delete(table_name, field_name, field_value)
}

func (db *DB) SelectUseSql(query_sql string, result_list *mysql_base.QueryResultList) error {
	return db.database.Query(query_sql, result_list)
}

func (db *DB) SelectRecordsCount(table_name string) (count int32, err error) {
	return db.database.SelectRecordsCount(table_name, "", nil)
}

func (db *DB) SelectRecordsCountByField(table_name, field_name string, field_value interface{}) (count int32, err error) {
	return db.database.SelectRecordsCount(table_name, field_name, field_value)
}

func (db *DB) Select(table_name string, field_name string, field_value interface{}, field_list []string, dest_list []interface{}) error {
	return db.database.SelectRecord(table_name, field_name, field_value, field_list, dest_list)
}

func (db *DB) SelectStar(table_name string, field_name string, field_value interface{}, dest_list []interface{}) error {
	return db.database.SelectRecord(table_name, field_name, field_value, nil, dest_list)
}

func (db *DB) SelectRecords(table_name string, field_name string, field_value interface{}, field_list []string, result_list *mysql_base.QueryResultList) error {
	return db.database.SelectRecords(table_name, field_name, field_value, field_list, result_list)
}

func (db *DB) SelectStarRecords(table_name string, field_name string, field_value interface{}, result_list *mysql_base.QueryResultList) error {
	return db.database.SelectRecords(table_name, field_name, field_value, nil, result_list)
}

func (db *DB) SelectAllRecords(table_name string, field_list []string, result_list *mysql_base.QueryResultList) error {
	return db.database.SelectRecords(table_name, "", nil, field_list, result_list)
}

func (db *DB) SelectStarAllRecords(table_name string, result_list *mysql_base.QueryResultList) error {
	return db.database.SelectRecords(table_name, "", nil, nil, result_list)
}

func (db *DB) SelectFieldNoKey(table_name string, field_name string, result_list *mysql_base.QueryResultList) error {
	return db.database.SelectRecords(table_name, "", nil, []string{field_name}, result_list)
}

func (db *DB) SelectRecordsCondition(table_name string, field_name string, field_value interface{}, sel_cond *mysql_base.SelectCondition, field_list []string, result_list *mysql_base.QueryResultList) error {
	return db.database.SelectRecordsCondition(table_name, field_name, field_value, sel_cond, field_list, result_list)
}

func (db *DB) NewTransaction() *Transaction {
	return db.op_mgr.NewTransaction()
}

func (db *DB) EndRun() bool {
	return atomic.CompareAndSwapInt32(&db.state, DB_STATE_RUNNING, DB_STATE_TO_END)
}

func (db *DB) IsEnd() bool {
	return atomic.LoadInt32(&db.state) == DB_STATE_NO_RUN
}

func (db *DB) Save() {
	db.op_mgr.Save()
}

func (db *DB) Run() {
	db.state = DB_STATE_RUNNING
	var last_save_time int32
	for {
		if atomic.CompareAndSwapInt32(&db.state, DB_STATE_TO_END, DB_STATE_NO_RUN) {
			break
		}

		now_time := int32(time.Now().Unix())
		if last_save_time == 0 {
			last_save_time = now_time
		} else if last_save_time > 0 && now_time-last_save_time >= int32(db.save_interval.Seconds()) {
			db.op_mgr.Save()
			last_save_time = now_time
		}
		time.Sleep(time.Second)
	}
	db.database.Close()
}

func (db *DB) GoRun() {
	go db.Run()
}
