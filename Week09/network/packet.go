package network

import (
	"encoding/binary"
	"github.com/valyala/bytebufferpool"
)

var packetEndian = binary.BigEndian // 网络字节序

// -----------------消息头------------------ | ------消息体------
// 消息长度，不含自身(uint32) | 协议ID(uint32) | 消息体(byte[])
type Packet struct {
	readIndex     uint32
	markReadIndex uint32
	writeIndex    uint32
	buff          *bytebufferpool.ByteBuffer
}

func NewPacket() *Packet {
	pkt := &Packet{}
	pkt.buff = bytebufferpool.Get()
	return pkt
}

func (p *Packet) GetReadIndex() uint32 {
	return p.readIndex
}

func (p *Packet) GetWriteIndex() uint32 {
	return p.writeIndex
}

func (p *Packet) readableData() []byte {
	return p.buff.B[p.readIndex:p.writeIndex]
}

func (p *Packet) data() []byte {
	return p.buff.B[0:p.writeIndex]
}

func (p *Packet) Length() uint32 {
	return p.writeIndex
}

func (p *Packet) ReadableBytes() uint32 {
	return p.writeIndex - p.readIndex
}

func (p *Packet) SetWriteIndex(index uint32) {
	p.writeIndex = index
}

// 是否有数据可读
func (p *Packet) Readable() bool {
	return p.readIndex < p.writeIndex
}

func (p *Packet) Release() {
	bytebufferpool.Put(p.buff)
	p.buff = nil
}

func (p *Packet) WriteByte(b byte) {
	p.buff.WriteByte(b)
	p.writeIndex += 1
}

func (p *Packet) ReadByte() (v byte) {
	v = p.buff.B[p.readIndex]
	p.readIndex += 1
	return
}

func (p *Packet) GetByte(index uint32) (v byte) {
	v = p.buff.B[index]
	return
}

func (p *Packet) SetByte(index uint32, v byte) {
	p.buff.B[index] = v
}

func (p *Packet) WriteBool(v bool) {
	if v {
		p.WriteByte(1)
	} else {
		p.WriteByte(0)
	}
}

func (p *Packet) ReadBool() (v bool) {
	return p.ReadByte() != 0
}

func (p *Packet) WriteInt16(v int16) {
	p.buff.WriteByte(byte(v >> 8))
	p.buff.WriteByte(byte(v))
	p.writeIndex += 2
}

func (p *Packet) ReadInt16() (v int16) {
	v = int16(packetEndian.Uint16(p.buff.B[p.readIndex : p.readIndex+2]))
	p.readIndex += 2
	return
}

func (p *Packet) WriteUint32(v uint32) {
	p.buff.WriteByte(byte(v >> 24))
	p.buff.WriteByte(byte(v >> 16))
	p.buff.WriteByte(byte(v >> 8))
	p.buff.WriteByte(byte(v))

	p.writeIndex += 4
}

func (p *Packet) ReadUint32() (v uint32) {
	v = packetEndian.Uint32(p.buff.B[p.readIndex : p.readIndex+4])
	p.readIndex += 4
	return
}

func (p *Packet) WriteInt32(v int32) {
	p.buff.WriteByte(byte(v >> 24))
	p.buff.WriteByte(byte(v >> 16))
	p.buff.WriteByte(byte(v >> 8))
	p.buff.WriteByte(byte(v))
	p.writeIndex += 4
}

func (p *Packet) ReadInt32() (v int32) {
	v = int32(packetEndian.Uint32(p.buff.B[p.readIndex : p.readIndex+4]))
	p.readIndex += 4
	return
}

func (p *Packet) WriteInt64(v int64) {
	p.buff.WriteByte(byte(v >> 56))
	p.buff.WriteByte(byte(v >> 48))
	p.buff.WriteByte(byte(v >> 40))
	p.buff.WriteByte(byte(v >> 32))
	p.buff.WriteByte(byte(v >> 24))
	p.buff.WriteByte(byte(v >> 16))
	p.buff.WriteByte(byte(v >> 8))
	p.buff.WriteByte(byte(v >> v))

	p.writeIndex += 8
}

func (p *Packet) ReadInt64() (v int64) {
	v = int64(packetEndian.Uint64(p.buff.B[p.readIndex : p.readIndex+8]))
	p.readIndex += 8
	return
}

func (p *Packet) WriteString(str string) {
	bs := []byte(str)
	len := uint32(len(bs))
	p.WriteUint32(len + 1)
	if len > 0 {
		p.buff.Write(bs)
	}
	p.writeIndex += len

	// 补个空字符在最后面
	p.WriteByte(0)
}

func (p *Packet) WriteStringBytes(bytes []byte) {
	len := uint32(len(bytes))
	p.WriteUint32(len + 1)
	if len > 0 {
		p.buff.Write(bytes)
	}
	p.writeIndex += len
	// 补个空字符在最后面
	p.WriteByte(0)
}

func (p *Packet) ReadString() (str string) {
	len := p.ReadUint32()
	str = string(p.buff.B[p.readIndex : p.readIndex+len-1])
	p.readIndex += len
	return
}

func (p *Packet) MarkReadIndex() {
	p.markReadIndex = p.readIndex
}

func (p *Packet) ResetReadIndex2Mark() {
	p.readIndex = p.markReadIndex
}

// 重置读写索引为0
func (p *Packet) ResetIndex() {
	p.readIndex = 0
	p.writeIndex = 0
}
