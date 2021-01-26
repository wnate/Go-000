package network

import (
	"encoding/binary"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"log"
	"net"
	"time"
)

func NewPacketConn(conn net.Conn) *PacketConn {
	c := &PacketConn{
		conn: conn,
	}
	return c
}

type PacketConn struct {
	conn net.Conn
}

func (c *PacketConn) SetRecvDeadline(deadline time.Time) error {
	return c.conn.SetReadDeadline(deadline)
}

func (c *PacketConn) SendPacket(pkt *Packet) error {
	defer func() {
		pkt.Release()
	}()
	len := pkt.ReadableBytes()
	log.Println("conn SendPacket start 1")
	err := binary.Write(c.conn, binary.BigEndian, &len)
	if err != nil {
		return err
	}
	log.Println("conn SendPacket start 2")
	err = writeAll(c.conn, pkt.readableData())
	if err != nil {
		return err
	}
	log.Println("conn SendPacket start 3")
	return nil
}

func (c *PacketConn) ReadPacket() (*Packet, error) {
	var (
		err    error
		pkt    *Packet
		pktLen uint32
	)
	err = binary.Read(c.conn, binary.BigEndian, &pktLen)
	if err != nil {
		return nil, err
	}
	if pktLen < 1 || pktLen > 9999 {
		return nil, errors.New(fmt.Sprintf("conn [%s] receive illegal packet len:%d, conn will closed",
			c, pktLen))
	}

	pkt = NewPacket()

	buf := make([]byte, pktLen)
	// 读取整个消息
	_, err = io.ReadFull(c.conn, buf[0:pktLen])

	pkt.buff.Write(buf)

	if err != nil {
		// 回收packet
		pkt.Release()
		return nil, err
	}
	pkt.SetWriteIndex(pktLen)

	return pkt, nil
}

func (c *PacketConn) Close() error {
	if tc, ok := c.conn.(*net.TCPConn); ok {
		tc.SetLinger(0)
	}
	c.conn.Close()
	return nil
}

func (c *PacketConn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *PacketConn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *PacketConn) String() string {
	return fmt.Sprintf("[%s >>> %s]", c.LocalAddr(), c.RemoteAddr())
}

func writeAll(conn io.Writer, data []byte) error {
	left := len(data)
	for left > 0 {
		n, err := conn.Write(data)
		if n == left && err == nil { // handle most common case first
			return nil
		}

		if n > 0 {
			data = data[n:]
			left -= n
		}

		if err != nil {
			return err
		}
	}
	return nil
}
