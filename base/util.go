package mysql_base

import (
	"log"
	"os"
)

func CopySrcValue2Dest(dest, src interface{}) bool {
	switch dt := dest.(type) {
	case *int8:
		d := dest.(*int8)
		r := src.(int8)
		*d = r
	case *int16:
		d := dest.(*int16)
		r := src.(int16)
		*d = r
	case *int32:
		d := dest.(*int32)
		r := src.(int32)
		*d = r
	case *int64:
		d := dest.(*int64)
		r := src.(int64)
		*d = r
	case *int:
		d := dest.(*int)
		r := src.(int)
		*d = r
	case *uint8:
		d := dest.(*uint8)
		r := src.(uint8)
		*d = r
	case *uint16:
		d := dest.(*uint16)
		r := src.(uint16)
		*d = r
	case *uint32:
		d := dest.(*uint32)
		r := src.(uint32)
		*d = r
	case *uint64:
		d := dest.(*uint64)
		r := src.(uint64)
		*d = r
	case *uint:
		d := dest.(*uint)
		r := src.(uint)
		*d = r
	case *bool:
		d := dest.(*bool)
		r := src.(bool)
		*d = r
	case *float32:
		d := dest.(*float32)
		r := src.(float32)
		*d = r
	case *float64:
		d := dest.(*float64)
		r := src.(float64)
		*d = r
	case *string:
		d := dest.(*string)
		r := src.(string)
		*d = r
	case *[]byte:
		d := dest.(*[]byte)
		r := src.([]byte)
		*d = r
	default:
		log.Printf("mysql_base: copy src value to unsupported dest type %v\n", dt)
		return false
	}
	return true
}

func CreateDirs(dest_path string) (err error) {
	if err = os.MkdirAll(dest_path, os.ModePerm); err != nil {
		log.Printf("创建目录结构%v错误 %v\n", dest_path, err.Error())
		return
	}
	if err = os.Chmod(dest_path, os.ModePerm); err != nil {
		log.Printf("修改目录%v权限错误 %v\n", dest_path, err.Error())
		return
	}
	return
}
