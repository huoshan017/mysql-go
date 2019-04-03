package mysql_base

type TableBase interface {
	Insert() bool
	Update() bool
	Delete() bool
	Name() string
}

type TableManager struct {
	tables_array []TableBase
	tables_map   map[string]TableBase
}

func (this *TableManager) Add(table TableBase) bool {
	return true
}

func (this *TableManager) Remove(table TableBase) bool {
	return true
}
