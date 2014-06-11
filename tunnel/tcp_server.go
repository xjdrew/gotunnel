//
//   date  : 2014-06-04
//   author: xjdrew
//

package tunnel

import (
	"net"

	"errors"
)

type TcpServer struct {
	addr string
	ln   *net.TCPListener
}

func (self *TcpServer) accept() (conn *net.TCPConn, err error) {
	conn, err = self.ln.AcceptTCP()
	return
}

func (self *TcpServer) buildListener() error {
	if self.ln != nil {
		return errors.New("server has started")
	}

	laddr, err := net.ResolveTCPAddr("tcp", self.addr)
	if err != nil {
		Error("resolve local addr failed:%s", err.Error())
		return err
	}

	ln, err := net.ListenTCP("tcp", laddr)
	if err != nil {
		Error("build listener failed:%s", err.Error())
		return err
	}

	Info("listen %s", self.addr)
	self.ln = ln
	return nil
}

func (self *TcpServer) closeListener() {
	if self.ln != nil {
		self.ln.Close()
	}
}
