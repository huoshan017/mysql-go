package mysql_proxy

import (
	"encoding/gob"
	"errors"
	"fmt"
	"log"
	"sync/atomic"
	"time"

	"github.com/huoshan017/mysql-go/proxy/common"
)

const (
	RPC_CLIENT_STATE_NONE = iota
	RPC_CLIENT_STATE_CONNECTING
	RPC_CLIENT_STATE_CONNECTED
	RPC_CLIENT_STATE_DISCONNECT
)

const (
	PING_INTERVAL = 5
)

func RegisterUserType(rcvr interface{}) {
	gob.Register(rcvr)
}

type client_inter interface {
	Call(string, interface{}, interface{}) error
	Close() error
}

type OnConnectFunc func(arg interface{})

type client struct {
	c          client_inter
	conn_type  int32
	state      int32 // 只在Run协程中修改
	addr       string
	on_connect OnConnectFunc
	to_close   int32
	ping_args  mysql_proxy_common.PingArgs
	ping_reply mysql_proxy_common.PongReply
}

func new_client() *client {
	return &client{}
}

func (this *client) ping() error {
	var err error
	if this.conn_type == mysql_proxy_common.CONNECTION_TYPE_ONLY_READ {
		err = this.Call("PingProc.Ping", &this.ping_args, &this.ping_reply)
	} else if this.conn_type == mysql_proxy_common.CONNECTION_TYPE_WRITE {
		c := this.c.(*mysql_proxy_common.ClientOnlyWrite)
		if c != nil {
			err = c.CallImmidiate("PingProc.Ping", &this.ping_args, &this.ping_reply)
		}
	} else {
		return errors.New(fmt.Sprintf("rpc unknown conn type %v", this.conn_type))
	}
	if err != nil {
		log.Printf("rpc client ping error[%v]\n", err.Error())
	}
	return err
}

func (this *client) SetOnConnect(on_connect OnConnectFunc) {
	this.on_connect = on_connect
}

func (this *client) Run() {
	for {
		to_close := atomic.LoadInt32(&this.to_close)
		if to_close > 0 {
			break
		}
		var err error
		if this.state == RPC_CLIENT_STATE_DISCONNECT {
			err = this.Dial(this.addr, this.conn_type)
			if err != nil {
				log.Printf("rpc client type %v reconnect addr[%v] failed\n", this.conn_type, this.addr)
			} else {
				log.Printf("rpc client type %v reconnect addr[%v] succeed\n", this.conn_type, this.addr)
			}
		} else {
			err = this.ping()
			if err != nil {
				atomic.CompareAndSwapInt32(&this.state, RPC_CLIENT_STATE_CONNECTED, RPC_CLIENT_STATE_DISCONNECT)
				log.Printf("rpc client type %v disconnected, ready to reconnect...\n", this.conn_type)
				time.Sleep(time.Second * PING_INTERVAL)
				continue
			}
		}
		time.Sleep(time.Second * PING_INTERVAL)
	}
}

func (this *client) RunBackground() {
	go func() {
		this.Run()
	}()
}

func (this *client) Dial(addr string, conn_type int32) error {
	var c client_inter
	var e error
	if conn_type == mysql_proxy_common.CONNECTION_TYPE_ONLY_READ {
		c, e = mysql_proxy_common.Dial("tcp", addr)
		if e != nil {
			log.Printf("rpc dial addr[%v] error[%v]\n", addr, e.Error())
			return e
		}
	} else if conn_type == mysql_proxy_common.CONNECTION_TYPE_WRITE {
		c, e = mysql_proxy_common.DialOnlyWrite("tcp", addr)
		if e != nil {
			log.Printf("rpc dial addr[%v] error[%v]\n", addr, e.Error())
			return e
		}
	} else {
		log.Printf("rpc dial connection invalid type: %v\n", conn_type)
		return fmt.Errorf("proxy client connection type %v invalid", conn_type)
	}
	this.c = c
	this.conn_type = conn_type
	this.state = RPC_CLIENT_STATE_CONNECTED
	this.addr = addr
	this.to_close = 0
	if this.on_connect != nil {
		this.on_connect(this)
	}
	return nil
}

func (this *client) Call(method string, args interface{}, reply interface{}) error {
	if this.c == nil {
		return errors.New("not create rpc client")
	}
	err := this.c.Call(method, args, reply)
	return err
}

func (this *client) Close() {
	if this.c != nil {
		this.c.Close()
		this.c = nil
		atomic.StoreInt32(&this.to_close, 1)
	}
}

func (this *client) GetState() int32 {
	return this.state
}
