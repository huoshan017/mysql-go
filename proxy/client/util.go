package mysql_proxy

import (
	"log"
)

func _copy_reply_value_2_dest(dest, reply interface{}) bool {
	switch dt := dest.(type) {
	case *int8:
		d := dest.(*int8)
		r := reply.(int8)
		*d = r
	case *int16:
		d := dest.(*int16)
		r := reply.(int16)
		*d = r
	case *int32:
		d := dest.(*int32)
		r := reply.(int32)
		*d = r
	case *int64:
		d := dest.(*int64)
		r := reply.(int64)
		*d = r
	case *int:
		d := dest.(*int)
		r := reply.(int)
		*d = r
	case *uint8:
		d := dest.(*uint8)
		r := reply.(uint8)
		*d = r
	case *uint16:
		d := dest.(*uint16)
		r := reply.(uint16)
		*d = r
	case *uint32:
		d := dest.(*uint32)
		r := reply.(uint32)
		*d = r
	case *uint64:
		d := dest.(*uint64)
		r := reply.(uint64)
		*d = r
	case *uint:
		d := dest.(*uint)
		r := reply.(uint)
		*d = r
	case *bool:
		d := dest.(*bool)
		r := reply.(bool)
		*d = r
	case *float32:
		d := dest.(*float32)
		r := reply.(float32)
		*d = r
	case *float64:
		d := dest.(*float64)
		r := reply.(float64)
		*d = r
	case *string:
		d := dest.(*string)
		r := reply.(string)
		*d = r
	case *[]byte:
		d := dest.(*[]byte)
		r := reply.([]byte)
		*d = r
	default:
		log.Printf("mysql-client: copy reply value to unsupported dest type %v\n", dt)
		return false
	}
	return true
}
