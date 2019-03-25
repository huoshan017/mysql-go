package mysql

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type Database struct {
	db *sql.DB
}

func (this *Database) open(dbhost, dbuser, dbpassword, dbname string) error {
	dsn := fmt.Sprintf("%v:%v@tcp(%v)/%v", dbuser, dbpassword, dbhost, dbname)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Printf("mysql:Database connect err %v\n", err.Error())
		return err
	}
	this.db = db
	log.Printf("mysql:Database connected db %v/%v with user %v\n", dbhost, dbname, dbuser)
	return nil
}

func (this *Database) close() {
	if this.db == nil {
		log.Printf("mysql:Database close failed with null instance")
		return
	}
	err := this.db.Close()
	if err != nil {
		log.Printf("mysql:Database close err %v", err.Error())
	} else {
		log.Printf("mysql:Database closed")
	}
}

func (this *Database) set_max_life_time(d time.Duration) {
	this.db.SetConnMaxLifetime(d)
}

func (this *Database) set_max_idle_conns(conns int) {
	this.db.SetMaxIdleConns(conns)
}

func (this *Database) set_max_open_conns(conns int) {
	this.db.SetMaxOpenConns(conns)
}

func (this *Database) query(query_str string) bool {
	var rows *sql.Rows
	var err error
	rows, err = this.db.Query(query_str)
	defer rows.Close()
	if err != nil {
		return false
	}
	return true
}

func (this *Database) query_with(query_str string, args ...interface{}) bool {
	rows, err := this.db.Query(query_str, args)
	defer rows.Close()
	if err != nil {
		return false
	}
	return true
}

func (this *Database) query_one(query_str string) bool {
	row := this.db.QueryRow(query_str)
	if row == nil {
		return false
	}
	return true
}

func (this *Database) query_one_with(query_str string, args ...interface{}) bool {
	row := this.db.QueryRow(query_str, args)
	if row == nil {
		return false
	}
	return true
}

func (this *Database) exec(query_str string) bool {
	res, err := this.db.Exec(query_str)
	if err != nil {
		log.Printf("mysql:Database exec err %v", err.Error())
		return false
	}
	res.LastInsertId()
	res.RowsAffected()
	return true
}
