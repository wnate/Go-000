package main

import (
	"github.com/wnate/Go-000/tree/main/Week09/network"
	"log"
)

func main() {
	svr := NewDemoServer(":9999")
	if svr.Start() != nil {
		log.Fatal("start server err")
	}

	select {}
}

func NewDemoServer(listenAddr string) *DemoServer {
	s := &DemoServer{
		network.NewTcpServer(listenAddr),
	}
	s.SetSessionEventHandler(s)
	return s
}

type DemoServer struct {
	*network.TCPServer
}

func (s *DemoServer) OnRecvPacket(session *network.Session, pkt *network.Packet) {
	log.Panicln("receive msg")
	// todo 根据具体的消息协议号做不同的handle
}

func (s *DemoServer) OnOpen(session *network.Session) error {
	log.Printf("new connection comming...%s\n", session.StrId())

	return nil
}

func (s *DemoServer) OnClose(session *network.Session) {
	log.Printf("connection close...%s\n", session.StrId())
}
