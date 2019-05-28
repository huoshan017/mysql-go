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

type ClientInter interface {
	Call(string, interface{}, interface{}) error
	Close() error
}

type OnConnectFunc func(arg interface{})

type Client struct {
	c          ClientInter
	conn_type  int32
	state      int32 // 只在Run协程中修改
	addr       string
	on_connect OnConnectFunc
	to_close   int32
	ping_args  mysql_proxy_common.PingArgs
	ping_reply mysql_proxy_common.PongReply
}

func NewClient() *Client {
	client := &Client{}
	return client
}

func (this *Client) ping() error {
	var err error
	if this.conn_type == mysql_proxy_common.CONNECTION_TYPE_ONLY_READ {
		err = this.Call("PingProc.Ping", &this.ping_args, &this.ping_reply)
	} else if this.conn_type == mysql_proxy_common.CONNECTION_TYPE_WRITE {
		client := this.c.(*mysql_proxy_common.ClientOnlyWrite)
		if client != nil {
			err = client.CallImmidiate("PingProc.Ping", &this.ping_args, &this.ping_reply)
		}
	} else {
		return errors.New(fmt.Sprintf("rpc unknown conn type %v", this.conn_type))
	}
	if err != nil {
		log.Printf("rpc client ping error[%v]\n", err.Error())
	}
	return err
}

func (this *Client) SetOnConnect(on_connect OnConnectFunc) {
	this.on_connect = on_connect
}

func (this *Client) Run() {
	go func() {
		for {
			to_close := atomic.LoadInt32(&this.to_close)
			if to_close > 0 {
				break
			}
			if this.state == RPC_CLIENT_STATE_DISCONNECT {
				if !this.Dial(this.addr, this.conn_type) {
					log.Printf("rpc client type %v reconnect addr[%v] failed\n", this.conn_type, this.addr)
				} else {
					log.Printf("rpc client type %v reconnect addr[%v] succeed\n", this.conn_type, this.addr)
				}
			} else {
				err := this.ping()
				if err != nil {
					atomic.CompareAndSwapInt32(&this.state, RPC_CLIENT_STATE_CONNECTED, RPC_CLIENT_STATE_DISCONNECT)
					log.Printf("rpc client type %v disconnected, ready to reconnect...\n", this.conn_type)
					time.Sleep(time.Second * PING_INTERVAL)
					continue
				}
			}
			time.Sleep(time.Second * PING_INTERVAL)
		}
	}()
}

func (this *Client) Dial(addr string, conn_type int32) bool {
	var c ClientInter
	var e error
	if conn_type == mysql_proxy_common.CONNECTION_TYPE_ONLY_READ {
		c, e = mysql_proxy_common.Dial("tcp", addr)
		if e != nil {
			log.Printf("rpc dial addr[%v] error[%v]\n", addr, e.Error())
			return false
		}
	} else if conn_type == mysql_proxy_common.CONNECTION_TYPE_WRITE {
		c, e = mysql_proxy_common.DialOnlyWrite("tcp", addr)
		if e != nil {
			log.Printf("rpc dial addr[%v] error[%v]\n", addr, e.Error())
			return false
		}
	} else {
		log.Printf("rpc dial connection invalid type: %v\n", conn_type)
		return false
	}
	this.c = c
	this.conn_type = conn_type
	this.state = RPC_CLIENT_STATE_CONNECTED
	this.addr = addr
	this.to_close = 0
	if this.on_connect != nil {
		this.on_connect(this)
	}
	return true
}

func (this *Client) Call(method string, args interface{}, reply interface{}) error {
	if this.c == nil {
		return errors.New("not create rpc client")
	}
	err := this.c.Call(method, args, reply)
	return err
}

func (this *Client) Close() {
	if this.c != nil {
		this.c.Close()
		this.c = nil
		atomic.StoreInt32(&this.to_close, 1)
	}
}

func (this *Client) GetState() int32 {
	return this.state
}
