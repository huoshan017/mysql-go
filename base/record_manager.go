package mysql_base

type TableOperation interface {
	Insert()
	Update()
	Delete()
}

type RecordsManager struct {
}
