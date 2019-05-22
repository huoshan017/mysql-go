package mysql_proxy

import (
	"encoding/gob"
	"errors"
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

type OnConnectFunc func(arg interface{})

type ClientInter interface {
	Call(string, interface{}, interface{}) error
	Close() error
}

type Client struct {
	c          ClientInter
	conn_type  int32
	state      int32 // 只在Run协程中修改
	addr       string
	on_connect OnConnectFunc
	to_close   int32
}

func NewClient() *Client {
	client := &Client{}
	return client
}

type PingArgs struct {
}

type PongReply struct {
}

func (this *Client) ping() error {
	args := &PingArgs{}
	reply := &PongReply{}
	err := this.Call("PingProc.Ping", args, reply)
	if err != nil {
		log.Printf("RPC client ping error[%v]\n", err.Error())
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
					log.Printf("RPC reconnect addr[%v] failed\n", this.addr)
				} else {
					log.Printf("RPC reconnect addr[%v] succeed\n", this.addr)
				}
			} else {
				err := this.ping()
				if err != nil {
					atomic.CompareAndSwapInt32(&this.state, RPC_CLIENT_STATE_CONNECTED, RPC_CLIENT_STATE_DISCONNECT)
					log.Printf("RPC connection disconnected, ready to reconnect...\n")
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
			log.Printf("RPC Dial addr[%v] error[%v]\n", addr, e.Error())
			return false
		}
	} else if conn_type == mysql_proxy_common.CONNECTION_TYPE_WRITE {
		c, e = mysql_proxy_common.DialOnlyWrite("tcp", addr)
		if e != nil {
			log.Printf("RPC Dial addr[%v] error[%v]\n", addr, e.Error())
			return false
		}
	} else {
		log.Printf("RPC Dial connection invalid type: %v\n", conn_type)
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

func RegisterUserType(rcvr interface{}) {
	gob.Register(rcvr)
}
