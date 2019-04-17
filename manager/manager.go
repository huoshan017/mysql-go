package mysql_manager

import (
	"log"
	"time"

	"github.com/huoshan017/mysql-go/base"
	"github.com/huoshan017/mysql-go/generator"
)

const (
	DEFAULT_CONN_MAX_LIFE_SECONDS = time.Second * 5
	DEFAULT_SAVE_INTERVAL_TIME    = time.Minute * 5
)

type DB struct {
	config_loader mysql_generator.ConfigLoader
	database      mysql_base.Database
	db_op_manager mysql_base.DBOperateManager
}

func (this *DB) LoadConfig(config_path string) bool {
	if !this.config_loader.Load(config_path) {
		log.Printf("load config %v failed\n", config_path)
		return false
	}
	return true
}

func (this *DB) Connect(dbhost, dbuser, dbpassword, dbname string) bool {
	err := this.database.Open(dbhost, dbuser, dbpassword, this.config_loader.DBPkg)
	if err != nil {
		log.Printf("open database err %v\n", err.Error())
		return false
	}
	this.database.SetMaxLifeTime(DEFAULT_CONN_MAX_LIFE_SECONDS)
	if this.config_loader.Tables != nil {
		for _, t := range this.config_loader.Tables {
			if !this.database.LoadTable(t) {
				log.Printf("load table %v config failed\n", t.Name)
				return false
			}
		}
	}
	this.db_op_manager.Init(&this.database)
	return true
}

func (this *DB) SetConnectLifeTime(d time.Duration) {
	this.database.SetMaxLifeTime(d)
}

func (this *DB) Close() {
	this.database.Close()
}

func (this *DB) Save() {
	this.db_op_manager.Save()
}

func (this *DB) Run() {
	go func() {
		this.db_op_manager.Save()
		time.Sleep(DEFAULT_SAVE_INTERVAL_TIME)
	}()
}
