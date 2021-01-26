package network

import (
	"context"
	"github.com/orcaman/concurrent-map"
	"log"
	"net"
	"sync"
)

func NewTcpServer(listenAddr string, opt ...TcpOption) *TCPServer {
	opts := NewDefaultTcpOptions()
	for _, o := range opt {
		if o == nil {
			continue
		}
		o(opts)
	}
	s := &TCPServer{
		addr:     listenAddr,
		sessions: cmap.New(),
		opts:     opts,
	}
	return s
}

type TCPServer struct {
	addr         string
	ln           net.Listener
	eventHandler SessionEventHandler
	opts         *TcpOptions
	sessions     cmap.ConcurrentMap
	stopFlag     bool
	stopOnce     sync.Once
}

func (s *TCPServer) Start() (err error) {

	s.ln, err = net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}

	go func() {

		for {
			closeChan := make(chan struct{})
			go s.loopAccept(closeChan)

			select {
			case <-closeChan:
				if s.stopFlag {
					return
				}
			}
		}

	}()

	log.Printf("tcp server[%s] started \n", s.addr)

	return
}

func (s *TCPServer) loopAccept(closeCh chan struct{}) {
	defer func() {
		close(closeCh) // 通知退出
	}()

	ctx, cancel := context.WithCancel(context.Background())

	for {
		conn, err := s.ln.Accept()

		if err != nil {
			if s.stopFlag {
				cancel()
			}
			continue
		}

		go s.handleNewConn(ctx, conn)
	}
}

func (s *TCPServer) handleNewConn(ctx context.Context, conn net.Conn) {
	// 设置socket参数
	tcpConn := conn.(*net.TCPConn)
	//tcpConn.SetWriteBuffer(s.opts.ConnWriteBuffSize)
	//tcpConn.SetReadBuffer(s.opts.ConnReadBuffSize)
	tcpConn.SetNoDelay(true)

	session := NewSession(conn)
	session.SetEventHandler(s)

	s.sessions.Set(session.StrId(), session)

	session.StartServe(ctx)
}

func (s *TCPServer) SetSessionEventHandler(eventHandler SessionEventHandler) {
	s.eventHandler = eventHandler
}

func (s *TCPServer) GetSession(sid string) *Session {
	obj, exist := s.sessions.Get(sid)
	if !exist {
		return nil
	}
	return obj.(*Session)
}

func (s *TCPServer) Stop() {
	s.stopOnce.Do(func() {
		s.stopFlag = true
		s.ln.Close()
		log.Printf("tcp server[%s] stopped \n", s.addr)
	})
}

func (s *TCPServer) OnOpen(session *Session) error {
	if s.eventHandler != nil {
		return s.eventHandler.OnOpen(session)
	}
	return nil
}

func (s *TCPServer) OnRecvPacket(session *Session, pkt *Packet) {
	if s.eventHandler != nil {
		s.eventHandler.OnRecvPacket(session, pkt)
	}
}

func (s *TCPServer) OnClose(session *Session) {
	s.sessions.Remove(session.strId)

	if s.eventHandler != nil {
		s.eventHandler.OnClose(session)
	}
}
