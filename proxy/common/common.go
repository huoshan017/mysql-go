package mysql_proxy_common

const (
	CONNECTION_TYPE_NONE      = iota
	CONNECTION_TYPE_ONLY_READ = 1
	CONNECTION_TYPE_WRITE     = 2
)

type PingArgs struct {
}

type PongReply struct {
}

type ArgsHead struct {
	db_host_id    int32
	db_host_alias string
	db_name       string
}

func (this *ArgsHead) GetDBHostId() int32 {
	return this.db_host_id
}

func (this *ArgsHead) GetDBHostAlias() string {
	return this.db_host_alias
}

func (this *ArgsHead) GetDBName() string {
	return this.db_name
}

// select
type SelectArgs struct {
	Head             *ArgsHead
	TableName        string
	WhereFieldName   string
	WhereFieldValue  interface{}
	SelectFieldNames []string
}

type SelectReply struct {
	Result []interface{}
}

// select records
type SelectRecordsArgs struct {
	Head             *ArgsHead
	TableName        string
	WhereFieldName   string
	WhereFieldValue  interface{}
	SelectFieldNames []string
}

type SelectRecordsReply struct {
	ResultList [][]interface{}
}

// select all records
type SelectAllRecordsArgs struct {
	Head             *ArgsHead
	TableName        string
	SelectFieldNames []string
}

type SelectAllRecordsReply struct {
	ResultList [][]interface{}
}

// select field
type SelectFieldArgs struct {
	Head            *ArgsHead
	TableName       string
	SelectFieldName string
}

type SelectFieldReply struct {
	ResultList []interface{}
}

// select records order by
type SelectRecordsOrderbyArgs struct {
	Head             *ArgsHead
	TableName        string
	WhereFieldName   string
	WhereFieldValue  interface{}
	SelectFieldNames []string
	Orderby          string
	Desc             bool
	Offset           int
	Limit            int
}

type SelectRecordsOrderbyReply struct {
	ResultList [][]interface{}
}
