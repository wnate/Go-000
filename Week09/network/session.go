package network

import (
	"context"
	"log"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

var sessionId uint32

type SessionPool interface {
	GetSession(sid string) *Session
}

func NewSession(conn net.Conn) *Session {
	id := newSessionId()
	strId := strconv.Itoa((int)(id))
	s := &Session{
		id:         id,
		strId:      strId,
		conn:       NewPacketConn(conn),
		inMsgCh:    make(chan *Packet, 100),
		outMsgCh:   make(chan *Packet, 100),
		cronPeriod: 1 * time.Second,
	}
	return s
}

func newSessionId() uint32 {
	return atomic.AddUint32(&sessionId, 1)
}

type Session struct {
	conn               *PacketConn
	inMsgCh            chan *Packet
	outMsgCh           chan *Packet
	handler            SessionEventHandler
	closeOnce          sync.Once
	closeFlag          bool
	id                 uint32
	strId              string
	cronPeriod         time.Duration
	CronCounter        uint16
	HandledPacketNum   int       // 已处理的包数
	LastRecvPacketTime time.Time // 最近一次收到消息的时间，用来检查客户端是否掉线了
}

func (s *Session) SetEventHandler(handler SessionEventHandler) {
	s.handler = handler
}

func (s *Session) SetCronPeriod(cronPeriod time.Duration) {
	s.cronPeriod = cronPeriod
}

func (s *Session) StartServeAndMonitor(ctx context.Context, closeChan chan struct{}) {
	defer func() {
		if closeChan != nil {
			close(closeChan)
		}
	}()
	s.StartServe(ctx)
}

func (s *Session) StartServe(ctx context.Context) {
	subCtx, cancel := context.WithCancel(ctx)
	defer func() {
		if r := recover(); r != nil {
			log.Println("session StartServe crashed", r)
		}
		cancel()
	}()

	defer s.Close()

	var (
		inMsg *Packet
		err   error
	)

	if s.handler != nil {
		err = s.handler.OnOpen(s)
		if err != nil {
			return
		}
	}

	subErrChan := make(chan error)

	go s.loopRead(subCtx, subErrChan)

	go s.loopWrite(subCtx)

	for {
		select {
		case <-ctx.Done():
			return
		case err = <-subErrChan:
			s.conn.Close()
			return
		case inMsg = <-s.inMsgCh:
			s.recvPacket(inMsg)
		}
	}
}

func (s *Session) recvPacket(pkt *Packet) {
	defer func() {
		if pkt != nil {
			// 回收这个包
			pkt.Release()
		}
	}()
	if s.handler != nil {
		s.handler.OnRecvPacket(s, pkt)
	}
}

func (s *Session) loopRead(ctx context.Context, errChan chan<- error) {
	var (
		pkt *Packet
		err error
	)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		pkt, err = s.conn.ReadPacket()
		if err != nil {
			errChan <- err
			return
		}
		s.inMsgCh <- pkt
		s.LastRecvPacketTime = time.Now()
	}

}

func (s *Session) loopWrite(ctx context.Context) {
	defer func() {
		// log.NetLogger.Infof("session[id=%s in=%d out=%d] loopWrite coroutine exit", s.strId, len(s.inMsgCh), len(s.outMsgCh))
	}()
	var (
		pkt *Packet
		ok  bool
	)
	for {
		select {
		case <-ctx.Done():
			return
		case pkt, ok = <-s.outMsgCh:
			if !ok {
				return
			}
			err := s.conn.SendPacket(pkt)
			if err != nil {
				return
			}
		}
	}
}

func (s *Session) Close() {
	s.closeOnce.Do(func() {
		s.conn.Close()
		if s.handler != nil {
			s.handler.OnClose(s)
		}
		s.clear()
	})
}

func (s *Session) SendPacket(pkt *Packet) {
	if s.closeFlag {
		pkt.Release()
		return
	}
	s.outMsgCh <- pkt
}

func (s *Session) Id() uint32 {
	return s.id
}

func (s *Session) StrId() string {
	return s.strId
}

func (s *Session) String() string {
	return s.conn.String()
}

func (s *Session) GetCurrentReadQSize() int {
	return len(s.inMsgCh)
}

func (s *Session) GetCurrentWriteQSize() int {
	return len(s.outMsgCh)
}

func (s *Session) clear() {
	s.closeFlag = true
	close(s.outMsgCh)
	close(s.inMsgCh)
	for m := range s.outMsgCh {
		m.Release()
	}
	for m := range s.inMsgCh {
		m.Release()
	}
}
