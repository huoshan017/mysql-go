package mysql_proxy_common

const (
	CONNECTION_TYPE_NONE      = iota
	CONNECTION_TYPE_ONLY_READ = 1
	CONNECTION_TYPE_WRITE     = 2
)
