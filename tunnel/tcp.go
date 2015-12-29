//
//   date  : 2015-12-28
//   author: xjdrew
//

package tunnel

import (
	"net"
	"time"
)

type TcpListener struct {
	*net.TCPListener
}

func (l *TcpListener) Accept() (net.Conn, error) {
	conn, err := l.TCPListener.AcceptTCP()
	if err != nil {
		return nil, err
	}
	conn.SetKeepAlive(true)
	conn.SetKeepAlivePeriod(TunnelKeepAlivePeriod)
	return conn, err
}

// create a tcp listener for server
func newTcpListener(laddr string) (net.Listener, error) {
	ln, err := net.Listen("tcp", laddr)
	if err != nil {
		return nil, err
	}
	tl := TcpListener{ln.(*net.TCPListener)}
	return &tl, nil
}

// for client
func dialTcp(raddr string) (net.Conn, error) {
	conn, err := net.DialTimeout("tcp", raddr, 5*time.Second)
	if err != nil {
		return nil, err
	}
	tcpConn := conn.(*net.TCPConn)
	tcpConn.SetKeepAlive(true)
	tcpConn.SetKeepAlivePeriod(TunnelKeepAlivePeriod)
	return tcpConn, nil
}
