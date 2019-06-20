package mysql_base

import (
	"errors"
)

var (
	ErrArgumentInvalid       = errors.New("argument invalid")
	ErrQueryResultEmpty      = errors.New("query row result is empty")
	ErrPrimaryFieldNotDefine = errors.New("primary field not defined")
	ErrInternal              = errors.New("internal error")
	ErrNoRows                = errors.New("no rows select")
)
