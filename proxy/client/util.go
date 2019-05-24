package mysql_proxy

import (
	"log"
)

func _copy_reply_value_2_dest(dest, reply interface{}) bool {
	switch dt := dest.(type) {
	case *int8:
		d := dest.(*int8)
		r := reply.(*int8)
		if r == nil {
			log.Printf("mysql-client: copy reply value transfer to *int8 failed\n", dt)
			return false
		}
		*d = *r
	case *int16:
		d := dest.(*int16)
		r := reply.(*int16)
		if r == nil {
			log.Printf("mysql-client: copy reply value transfer to *int16 failed\n", dt)
			return false
		}
		*d = *r
	case *int32:
		d := dest.(*int32)
		r := reply.(*int32)
		if r == nil {
			log.Printf("mysql-client: copy reply value transfer to *int32 failed\n", dt)
			return false
		}
		*d = *r
	case *int64:
		d := dest.(*int64)
		r := reply.(*int64)
		if r == nil {
			log.Printf("mysql-client: copy reply value transfer to *int64 failed\n", dt)
			return false
		}
		*d = *r
	case *int:
		d := dest.(*int)
		r := reply.(*int)
		if r == nil {
			log.Printf("mysql-client: copy reply value transfer to *int failed\n", dt)
			return false
		}
		*d = *r
	case *uint8:
		d := dest.(*uint8)
		r := reply.(*uint8)
		if r == nil {
			log.Printf("mysql-client: copy reply value transfer to *uint8 failed\n", dt)
			return false
		}
		*d = *r
	case *uint16:
		d := dest.(*uint16)
		r := reply.(*uint16)
		if r == nil {
			log.Printf("mysql-client: copy reply value transfer to *uint16 failed\n", dt)
			return false
		}
		*d = *r
	case *uint32:
		d := dest.(*uint32)
		r := reply.(*uint32)
		if r == nil {
			log.Printf("mysql-client: copy reply value transfer to *uint32 failed\n", dt)
			return false
		}
		*d = *r
	case *uint64:
		d := dest.(*uint64)
		r := reply.(*uint64)
		if r == nil {
			log.Printf("mysql-client: copy reply value transfer to *uint64 failed\n", dt)
			return false
		}
		*d = *r
	case *uint:
		d := dest.(*uint)
		r := reply.(*uint)
		if r == nil {
			log.Printf("mysql-client: copy reply value transfer to *uint failed\n", dt)
			return false
		}
		*d = *r
	case *bool:
		d := dest.(*bool)
		r := reply.(*bool)
		if r == nil {
			log.Printf("mysql-client: copy reply value transfer to *bool failed\n", dt)
			return false
		}
		*d = *r
	case *float32:
		d := dest.(*float32)
		r := reply.(*float32)
		if r == nil {
			log.Printf("mysql-client: copy reply value transfer to *float32 failed\n", dt)
			return false
		}
		*d = *r
	case *float64:
		d := dest.(*float64)
		r := reply.(*float64)
		if r == nil {
			log.Printf("mysql-client: copy reply value transfer to *float64 failed\n", dt)
			return false
		}
		*d = *r
	case *string:
		d := dest.(*string)
		r := reply.(*string)
		if r == nil {
			log.Printf("mysql-client: copy reply value transfer to *string failed\n", dt)
			return false
		}
		*d = *r
	case *[]byte:
		d := dest.(*[]byte)
		r := reply.([]byte)
		if r == nil {
			log.Printf("mysql-client: copy reply value transfer to []byte failed\n", dt)
			return false
		}
		*d = r
	default:
		log.Printf("mysql-client: copy reply value to unsupported dest type %v\n", dt)
		return false
	}
	return true
}
