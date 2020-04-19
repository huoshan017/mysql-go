package mysql_base

import (
	"database/sql"
	"fmt"
	"log"
	"time"
	//_ "github.com/go-sql-driver/mysql"
)

// QueryResultList ...
type QueryResultList struct {
	rows *sql.Rows
}

// CreateQueryResultList ...
func CreateQueryResultList(rows *sql.Rows) *QueryResultList {
	return &QueryResultList{
		rows: rows,
	}
}

// Init ...
func (q *QueryResultList) Init(rows *sql.Rows) {
	q.rows = rows
}

// Close ...
func (q *QueryResultList) Close() {
	if q.rows == nil {
		return
	}
	q.rows.Close()
}

// Get ...
func (q *QueryResultList) Get(dest ...interface{}) bool {
	if !q.rows.Next() {
		return false
	}
	err := q.rows.Scan(dest...)
	if err != nil {
		log.Printf("QueryResultList::Get with dest(%v) scan err %v\n", dest, err.Error())
		return false
	}
	return true
}

// Get2 ...
func (q *QueryResultList) Get2(dest []interface{}) bool {
	return q.Get(dest...)
}

// HasData ...
func (q *QueryResultList) HasData() bool {
	return q.rows.NextResultSet()
}

// Database ...
type Database struct {
	db *sql.DB
}

// Open ...
func (q *Database) Open(dbhost, dbuser, dbpassword, dbname string) error {
	dsn := fmt.Sprintf("%v:%v@tcp(%v)/%v", dbuser, dbpassword, dbhost, dbname)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}
	q.db = db
	return nil
}

// Close ...
func (q *Database) Close() {
	if q.db == nil {
		log.Printf("Database close failed with null instance\n")
		return
	}
	err := q.db.Close()
	if err != nil {
		log.Printf("Database close err %v\n", err.Error())
	} else {
		log.Printf("Database closed\n")
	}
}

// SetMaxLifeTime ...
func (q *Database) SetMaxLifeTime(d time.Duration) {
	q.db.SetConnMaxLifetime(d)
}

// SetMaxIdleConns ...
func (q *Database) SetMaxIdleConns(conns int) {
	q.db.SetMaxIdleConns(conns)
}

// SetMaxOpenConns ...
func (q *Database) SetMaxOpenConns(conns int) {
	q.db.SetMaxOpenConns(conns)
}

// Query ...
func (q *Database) Query(queryStr string, result *QueryResultList) error {
	rows, err := q.db.Query(queryStr)
	//defer rows.Close()
	if err != nil {
		return err
	}
	result.Init(rows)
	return nil
}

// QueryWith ...
func (q *Database) QueryWith(queryStr string, args []interface{}, result *QueryResultList) error {
	rows, err := q.db.Query(queryStr, args...)
	//defer rows.Close()
	if err != nil {
		return err
	}
	result.Init(rows)
	return nil
}

// QueryOne ...
func (q *Database) QueryOne(queryStr string, dest []interface{}) error {
	err := q.db.QueryRow(queryStr).Scan(dest...)
	if err == sql.ErrNoRows {
		err = ErrNoRows
	}
	return err
}

// QueryOneWith ...
func (q *Database) QueryOneWith(queryStr string, args []interface{}, dest []interface{}) error {
	err := q.db.QueryRow(queryStr, args...).Scan(dest...)
	if err == sql.ErrNoRows {
		err = ErrNoRows
	}
	return err
}

// QueryCount ...
func (q *Database) QueryCount(queryStr string) (count int32, err error) {
	err = q.db.QueryRow(queryStr).Scan(&count)
	return
}

// QueryCountWith ...
func (q *Database) QueryCountWith(queryStr string, arg interface{}) (count int32, err error) {
	err = q.db.QueryRow(queryStr, arg).Scan(&count)
	return
}

// HasRow ...
func (q *Database) HasRow(queryStr string) bool {
	row := q.db.QueryRow(queryStr)
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

func execResult(res sql.Result, lastInsertID, rowsAffected *int64) {
	var err error
	if lastInsertID != nil {
		*lastInsertID, err = res.LastInsertId()
		if err != nil {
			log.Printf("Database exec get last insert id err %v\n", err.Error())
		}
	}
	if rowsAffected != nil {
		*rowsAffected, err = res.RowsAffected()
		if err != nil {
			log.Printf("Database exe get rows affected err %v\n", err.Error())
		}
	}
}

// Exec ...
func (q *Database) Exec(queryStr string, lastInsertID, rowsAffected *int64) error {
	res, err := q.db.Exec(queryStr)
	if err != nil {
		log.Printf("Database exec query string(%v) err %v\n", queryStr, err.Error())
		return err
	}
	execResult(res, lastInsertID, rowsAffected)
	return nil
}

// ExecWith ...
func (q *Database) ExecWith(queryStr string, args []interface{}, lastInsertID, rowsAffected *int64) error {
	res, err := q.db.Exec(queryStr, args...)
	if err != nil {
		log.Printf("Database exec query string(%v) with args(%v) err %v\n", queryStr, args, err.Error())
		return err
	}
	execResult(res, lastInsertID, rowsAffected)
	return nil
}

// Prepare ...
func (q *Database) Prepare(queryStr string) *Stmt {
	stmt, err := q.db.Prepare(queryStr)
	if err != nil {
		log.Printf("Database Prepare query (%v) err %v\n", queryStr, err.Error())
		return nil
	}
	return CreateStmt(stmt)
}

// BeginProcedure ...
func (q *Database) BeginProcedure() *Procedure {
	tx, err := q.db.Begin()
	if err != nil {
		log.Printf("Database begin procedure err %v\n", err.Error())
		return nil
	}
	return CreateProcedure(tx)
}

// ----------------------------------- STMT -----------------------------------

// Stmt ...
type Stmt struct {
	stmt *sql.Stmt
}

// CreateStmt ...
func CreateStmt(stmt *sql.Stmt) *Stmt {
	return &Stmt{
		stmt: stmt,
	}
}

// Query ...
func (q *Stmt) Query(args []interface{}, result *QueryResultList) error {
	rows, err := q.stmt.Query(args...)
	defer rows.Close()
	if err != nil {
		log.Printf("Stmt query err %v\n", err.Error())
		return err
	}
	result.Init(rows)
	return nil
}

// QueryOne ...
func (q *Stmt) QueryOne(args []interface{}, dest []interface{}) error {
	row := q.stmt.QueryRow(args...)
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

// Exec ...
func (q *Stmt) Exec(args []interface{}, lastInsertID, rowsAffected *int64) error {
	res, err := q.stmt.Exec(args...)
	if err != nil {
		log.Printf("Stmt exec with args err %v\n", err.Error())
		return err
	}
	execResult(res, lastInsertID, rowsAffected)
	return nil
}

// -------------------------------- Procedure ---------------------------------

// Procedure ...
type Procedure struct {
	tx *sql.Tx
}

// CreateProcedure ...
func CreateProcedure(tx *sql.Tx) *Procedure {
	return &Procedure{
		tx: tx,
	}
}

// Query ...
func (p *Procedure) Query(queryStr string, result *QueryResultList) error {
	rows, err := p.tx.Query(queryStr)
	if err != nil {
		log.Printf("Procedure query(%v) err %v\n", queryStr, err.Error())
		return err
	}
	result.Init(rows)
	return nil
}

// QueryWith ...
func (p *Procedure) QueryWith(queryStr string, args []interface{}, result *QueryResultList) error {
	rows, err := p.tx.Query(queryStr, args...)
	if err != nil {
		log.Printf("Procedure query(%v) with args(%v) err %v\n", queryStr, args, err.Error())
		return err
	}
	result.Init(rows)
	return nil
}

// QueryOne ...
func (p *Procedure) QueryOne(queryStr string, dest []interface{}) error {
	err := p.tx.QueryRow(queryStr).Scan(dest...)
	if err != nil {
		log.Printf("Procedure query(%v) one row with args(%v) and scan err %v\n", queryStr, dest, err.Error())
		return err
	}
	return nil
}

// QueryOneWith ...
func (p *Procedure) QueryOneWith(queryStr string, args []interface{}, dest []interface{}) error {
	err := p.tx.QueryRow(queryStr, args...).Scan(dest...)
	if err != nil {
		log.Printf("Procedure query(%v) one row with args(%v) and scan err %v\n", queryStr, args, err.Error())
		return err
	}
	return nil
}

// Exec ...
func (p *Procedure) Exec(queryStr string, lastInsertID, rowsAffected *int64) error {
	res, err := p.tx.Exec(queryStr)
	if err != nil {
		log.Printf("Procedure exec(%v) with err %v\n", queryStr, err.Error())
		return err
	}
	execResult(res, lastInsertID, rowsAffected)
	return nil
}

// ExecWith ...
func (p *Procedure) ExecWith(queryStr string, args []interface{}, lastInsertID, rowsAffected *int64) error {
	res, err := p.tx.Exec(queryStr, args...)
	if err != nil {
		log.Printf("Procedure exec(%v) with args(%v) err %v\n", queryStr, args, err.Error())
		return err
	}
	execResult(res, lastInsertID, rowsAffected)
	return nil
}

// Commit ...
func (p *Procedure) Commit() error {
	err := p.tx.Commit()
	if err != nil {
		log.Printf("Procedure commit err %v\n", err.Error())
		return err
	}
	return nil
}

// Rollback ...
func (p *Procedure) Rollback() error {
	err := p.tx.Rollback()
	if err != nil {
		log.Printf("Procedure rollback err %v\n", err.Error())
		return err
	}
	return nil
}
