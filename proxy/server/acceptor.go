package main

import (
	"encoding/gob"
	"log"
	"net"
	"net/rpc"

	"github.com/huoshan017/mysql-go/proxy/common"
)

type PingProc struct {
}

func (this *PingProc) Ping(args *mysql_proxy_common.PingArgs, reply *mysql_proxy_common.PongReply) error {
	return nil
}

type Service struct {
	listener *net.TCPListener
}

func (this *Service) Register(rcvr interface{}) bool {
	err := rpc.Register(rcvr)
	if err != nil {
		log.Printf("rpc service register error[%v]\n", err.Error())
		return false
	}
	return true
}

func (this *Service) Listen(addr string) error {
	ping_proc := &PingProc{}
	this.Register(ping_proc)

	var address, _ = net.ResolveTCPAddr("tcp", addr)
	l, e := net.ListenTCP("tcp", address)
	if e != nil {
		return e
	}
	this.listener = l

	log.Printf("rpc service listen to %v\n", addr)
	return nil
}

func (this *Service) Serve() {
	server := mysql_proxy_common.NewServer()
	server.Accept(this.listener)
}

func (this *Service) Close() {
	this.listener.Close()
}

func RegisterUserType(rcvr interface{}) {
	gob.Register(rcvr)
}
