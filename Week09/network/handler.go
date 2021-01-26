package network

type SessionEventHandler interface {
	OnOpen(session *Session) error

	OnClose(session *Session)

	OnRecvPacket(session *Session, pkt *Packet)
}
