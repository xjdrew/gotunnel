//
//   date  : 2015-12-28
//   author: xjdrew
//

package tunnel

import (
	"net"
	"time"

	udt "github.com/xjdrew/go-udtwrapper/udt"
)

type UdtListener struct {
	net.Listener
}

func setUdtOpt(conn *udt.UDTConn) {
	conn.SetDeadline(time.Time{})
	conn.SetLinger(0)
}

func (l *UdtListener) Accept() (net.Conn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}
	setUdtOpt(conn.(*udt.UDTConn))
	return conn, err
}

// create a udt listener for server
func newUdtListener(laddr string) (net.Listener, error) {
	ln, err := udt.Listen("udt", laddr)
	if err != nil {
		return nil, err
	}
	ul := UdtListener{ln}
	return &ul, nil
}

// for client
func dialUdt(raddr string) (net.Conn, error) {
	conn, err := udt.Dial("udt", raddr)
	if err == nil {
		setUdtOpt(conn.(*udt.UDTConn))
	}
	return conn, err
}
