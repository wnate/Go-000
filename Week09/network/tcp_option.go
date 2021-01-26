package network

func NewDefaultTcpOptions() *TcpOptions {
	opts := &TcpOptions{
		ConnReadBuffSize:  1024 * 1024,
		ConnWriteBuffSize: 1024 * 1024,
	}
	return opts
}

type TcpOption func(*TcpOptions)

type TcpOptions struct {
	ConnReadBuffSize  int
	ConnWriteBuffSize int
}
