package mysql_base

import (
	"log"
)

type Row interface {
	Insert() bool
	Update() bool
	Delete() bool
}

type Table struct {
	Name() string
}

type TableManager struct {
	tables_array []TableBase
	tables_map   map[string]TableBase
	db           *Database
}

func (this *TableManager) Init(db *Database) {
	this.db = db
}

func (this *TableManager) Add(table TableBase) bool {
	if this.tables_map == nil {
		this.tables_map = make(map[string]TableBase)
	}
	name := table.Name()
	var t TableBase
	t = this.tables_map[name]
	if t != nil {
		log.Printf("TableManager::Add already has table %v", name)
		return false
	}
	if this.tables_array != nil {
		for _, t = range this.tables_array {
			if t.Name() == name {
				log.Printf("TableManager::Add already has table %v", name)
				return false
			}
		}
	}
	this.tables_array = append(this.tables_array, table)
	this.tables_map[name] = table
	return true
}

func (this *TableManager) RemoveByName(name string) bool {
	if this.tables_array == nil || this.tables_map == nil {
		return false
	}
	var found bool
	for i := 0; i < len(this.tables_array); i++ {
		t := this.tables_array[i]
		if t != nil && t.Name() == name {
			this.tables_array[i] = nil
			found = true
			break
		}
	}
	delete(this.tables_map, name)
	return found
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
