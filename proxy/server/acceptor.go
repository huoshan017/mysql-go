package main

import (
	"log"
	"net"
	"net/rpc"

	"github.com/huoshan017/mysql-go/proxy/client"
	"github.com/huoshan017/mysql-go/proxy/common"
)

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

type PingProc struct {
}

func (this *PingProc) Ping(args *mysql_proxy.PingArgs, reply *mysql_proxy.PongReply) error {
	return nil
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
