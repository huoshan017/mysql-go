package mysql_base

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var (
	ErrNoRows = errors.New("no rows select")
)

type QueryResultList struct {
	rows *sql.Rows
}

func CreateQueryResultList(rows *sql.Rows) *QueryResultList {
	return &QueryResultList{
		rows: rows,
	}
}

func (this *QueryResultList) Init(rows *sql.Rows) {
	this.rows = rows
}

func (this *QueryResultList) Close() {
	if this.rows == nil {
		return
	}
	this.rows.Close()
}

func (this *QueryResultList) Get(dest ...interface{}) bool {
	if !this.rows.Next() {
		return false
	}
	err := this.rows.Scan(dest...)
	if err != nil {
		log.Printf("QueryResultList::Get with dest(%v) scan err %v\n", dest, err.Error())
		return false
	}
	return true
}

func (this *QueryResultList) Get2(dest []interface{}) bool {
	return this.Get(dest...)
}

func (this *QueryResultList) HasData() bool {
	return this.rows.NextResultSet()
}

type Database struct {
	db *sql.DB
}

func (this *Database) Open(dbhost, dbuser, dbpassword, dbname string) error {
	dsn := fmt.Sprintf("%v:%v@tcp(%v)/%v", dbuser, dbpassword, dbhost, dbname)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}
	this.db = db
	return nil
}

func (this *Database) Close() {
	if this.db == nil {
		log.Printf("Database close failed with null instance\n")
		return
	}
	err := this.db.Close()
	if err != nil {
		log.Printf("Database close err %v\n", err.Error())
	} else {
		log.Printf("Database closed\n")
	}
}

func (this *Database) SetMaxLifeTime(d time.Duration) {
	this.db.SetConnMaxLifetime(d)
}

func (this *Database) SetMaxIdleConns(conns int) {
	this.db.SetMaxIdleConns(conns)
}

func (this *Database) SetMaxOpenConns(conns int) {
	this.db.SetMaxOpenConns(conns)
}

func (this *Database) Query(query_str string, result *QueryResultList) error {
	rows, err := this.db.Query(query_str)
	//defer rows.Close()
	if err != nil {
		return err
	}
	result.Init(rows)
	return nil
}

func (this *Database) QueryWith(query_str string, args []interface{}, result *QueryResultList) error {
	rows, err := this.db.Query(query_str, args...)
	//defer rows.Close()
	if err != nil {
		return err
	}
	result.Init(rows)
	return nil
}

func (this *Database) QueryOne(query_str string, dest []interface{}) error {
	err := this.db.QueryRow(query_str).Scan(dest...)
	if err == sql.ErrNoRows {
		err = ErrNoRows
	}
	return err
}

func (this *Database) QueryOneWith(query_str string, args []interface{}, dest []interface{}) error {
	err := this.db.QueryRow(query_str, args...).Scan(dest...)
	if err == sql.ErrNoRows {
		err = ErrNoRows
	}
	return err
}

func (this *Database) QueryCount(query_str string) (count int32, err error) {
	err = this.db.QueryRow(query_str).Scan(&count)
	return
}

func (this *Database) QueryCountWith(query_str string, arg interface{}) (count int32, err error) {
	err = this.db.QueryRow(query_str, arg).Scan(&count)
	return
}

func (this *Database) HasRow(query_str string) bool {
	row := this.db.QueryRow(query_str)
	if row == nil {
		return false
	}
	var dest interface{}
	err := row.Scan(dest)
	if err == sql.ErrNoRows {
		return false
	}
	return true
}

func _exec_result(res sql.Result, last_insert_id, rows_affected *int64) {
	var err error
	if last_insert_id != nil {
		*last_insert_id, err = res.LastInsertId()
		if err != nil {
			log.Printf("Database exec get last insert id err %v\n", err.Error())
		}
	}
	if rows_affected != nil {
		*rows_affected, err = res.RowsAffected()
		if err != nil {
			log.Printf("Database exe get rows affected err %v\n", err.Error())
		}
	}
}

func (this *Database) Exec(query_str string, last_insert_id, rows_affected *int64) error {
	res, err := this.db.Exec(query_str)
	if err != nil {
		log.Printf("Database exec query string(%v) err %v\n", query_str, err.Error())
		return err
	}
	_exec_result(res, last_insert_id, rows_affected)
	return nil
}

func (this *Database) ExecWith(query_str string, args []interface{}, last_insert_id, rows_affected *int64) error {
	res, err := this.db.Exec(query_str, args...)
	if err != nil {
		log.Printf("Database exec query string(%v) with args(%v) err %v\n", query_str, args, err.Error())
		return err
	}
	_exec_result(res, last_insert_id, rows_affected)
	return nil
}

func (this *Database) Prepare(query_str string) *Stmt {
	stmt, err := this.db.Prepare(query_str)
	if err != nil {
		log.Printf("Database Prepare query (%v) err %v\n", query_str, err.Error())
		return nil
	}
	return CreateStmt(stmt)
}

func (this *Database) BeginProcedure() *Procedure {
	tx, err := this.db.Begin()
	if err != nil {
		log.Printf("Database begin procedure err %v\n", err.Error())
		return nil
	}
	return CreateProcedure(tx)
}

// ----------------------------------- STMT -----------------------------------

type Stmt struct {
	stmt *sql.Stmt
}

func CreateStmt(stmt *sql.Stmt) *Stmt {
	return &Stmt{
		stmt: stmt,
	}
}

func (this *Stmt) Query(args []interface{}, result *QueryResultList) error {
	rows, err := this.stmt.Query(args...)
	defer rows.Close()
	if err != nil {
		log.Printf("Stmt query err %v\n", err.Error())
		return err
	}
	result.Init(rows)
	return nil
}

func (this *Stmt) QueryOne(args []interface{}, dest []interface{}) error {
	row := this.stmt.QueryRow(args...)
	if row == nil {
		//log.Printf("Stmt query one row get result empty\n")
		return ErrQueryResultEmpty
	}
	err := row.Scan(dest...)
	if err != nil {
		log.Printf("Stmt query one row and scan err %v\n", err.Error())
		return err
	}
	return nil
}

func (this *Stmt) Exec(args []interface{}, last_insert_id, rows_affected *int64) error {
	res, err := this.stmt.Exec(args...)
	if err != nil {
		log.Printf("Stmt exec with args err %v\n", err.Error())
		return err
	}
	_exec_result(res, last_insert_id, rows_affected)
	return nil
}

// -------------------------------- Procedure ---------------------------------
type Procedure struct {
	tx *sql.Tx
}

func CreateProcedure(tx *sql.Tx) *Procedure {
	return &Procedure{
		tx: tx,
	}
}

func (this *Procedure) Query(query_str string, result *QueryResultList) error {
	rows, err := this.tx.Query(query_str)
	if err != nil {
		log.Printf("Procedure query(%v) err %v\n", query_str, err.Error())
		return err
	}
	result.Init(rows)
	return nil
}

func (this *Procedure) QueryWith(query_str string, args []interface{}, result *QueryResultList) error {
	rows, err := this.tx.Query(query_str, args...)
	if err != nil {
		log.Printf("Procedure query(%v) with args(%v) err %v\n", query_str, args, err.Error())
		return err
	}
	result.Init(rows)
	return nil
}

func (this *Procedure) QueryOne(query_str string, dest []interface{}) error {
	err := this.tx.QueryRow(query_str).Scan(dest...)
	if err != nil {
		log.Printf("Procedure query(%v) one row with args(%v) and scan err %v\n", query_str, err.Error())
		return err
	}
	return nil
}

func (this *Procedure) QueryOneWith(query_str string, args []interface{}, dest []interface{}) error {
	err := this.tx.QueryRow(query_str, args...).Scan(dest...)
	if err != nil {
		log.Printf("Procedure query(%v) one row with args(%v) and scan err %v\n", query_str, args, err.Error())
		return err
	}
	return nil
}

func (this *Procedure) Exec(query_str string, last_insert_id, rows_affected *int64) error {
	res, err := this.tx.Exec(query_str)
	if err != nil {
		log.Printf("Procedure exec(%v) with err %v\n", query_str, err.Error())
		return err
	}
	_exec_result(res, last_insert_id, rows_affected)
	return nil
}

func (this *Procedure) ExecWith(query_str string, args []interface{}, last_insert_id, rows_affected *int64) error {
	res, err := this.tx.Exec(query_str, args...)
	if err != nil {
		log.Printf("Procedure exec(%v) with args(%v) err %v\n", query_str, args, err.Error())
		return err
	}
	_exec_result(res, last_insert_id, rows_affected)
	return nil
}

func (this *Procedure) Commit() error {
	err := this.tx.Commit()
	if err != nil {
		log.Printf("Procedure commit err %v\n", err.Error())
		return err
	}
	return nil
}

func (this *Procedure) Rollback() error {
	err := this.tx.Rollback()
	if err != nil {
		log.Printf("Procedure rollback err %v\n", err.Error())
		return err
	}
	return nil
}
