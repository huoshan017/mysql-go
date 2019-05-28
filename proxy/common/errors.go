package mysql_proxy_common

import (
	"errors"
	"log"
	"os"

	//"reflect"
	"runtime"
)

var ServerLogErr = Logger(log.New(os.Stderr, "[mysql-proxy-server] ", log.Ldate|log.Ltime|log.Lshortfile))
var ProxyLogErr = Logger(log.New(os.Stderr, "[mysql-proxy]", log.Ldate|log.Ltime|log.Lshortfile))

// Logger is used to log critical error messages.
type Logger interface {
	Print(v ...interface{})
}

// SetLogger is used to set the logger for critical errors.
// The initial logger is os.Stderr.
func SetServerLogger(logger Logger) error {
	if logger == nil {
		return errors.New("logger is nil")
	}
	ServerLogErr = logger
	return nil
}

func SetProxyLogger(logger Logger) error {
	if logger == nil {
		return errors.New("logger is nil")
	}
	ProxyLogErr = logger
	return nil
}

const (
	DEFAULT_CRITICAL_LOG_CALL_FRAME_NUM = 10
)

func OutputCriticalStack(logger Logger, err interface{}) {
	if logger == nil {
		return
	}
	logger.Print("<critical> %v", err)
	for i := 0; i < DEFAULT_CRITICAL_LOG_CALL_FRAME_NUM; i++ {
		funcName, file, line, ok := runtime.Caller(i)
		if ok {
			logger.Print("<stack>%v|%v|%v|%v}\n", i, runtime.FuncForPC(funcName).Name(), file, line)
		}
	}
}
