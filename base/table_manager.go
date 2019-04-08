package mysql_base

import (
	"log"
	"sync"
)

type TableBase interface {
	Insert() bool
	Update() bool
	Delete() bool
	Name() string
}

type TableManager struct {
	tables_map map[string]TableBase
	locker     sync.RWMutex
	db         *Database
}

func (this *TableManager) Init(db *Database) {
	this.tables_map = make(map[string]TableBase)
	this.db = db
}

func (this *TableManager) Add(table TableBase) bool {
	name := table.Name()
	var t TableBase
	t = this.tables_map[name]
	if t != nil {
		log.Printf("TableManager::Add already has table %v", name)
		return false
	}
	this.tables_map[name] = table
	return true
}

func (this *TableManager) RemoveByName(name string) bool {
	if this.tables_map == nil {
		return false
	}
	if _, o := this.tables_map[name]; !o {
		return false
	}
	delete(this.tables_map, name)
	return true
}

func (this *TableManager) Remove(table TableBase) bool {
	return this.RemoveByName(table.Name())
}

func (this *TableManager) Get(name string) TableBase {
	if this.tables_map == nil {
		return nil
	}
	return this.tables_map[name]
}

func (this *TableManager) Update() {

}
