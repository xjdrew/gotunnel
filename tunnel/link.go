//
//   date  : 2014-06-04
//   author: xjdrew
//

package tunnel

import (
	"bufio"
	"errors"
	"net"
)

var errPeerClosed = errors.New("errPeerClosed")

type Link struct {
	id   uint16
	conn *net.TCPConn
	err  error
}

func (self *Link) setError(err error) {
	if self.err != nil {
		return
	}
	self.err = err
}

// write data to peer
func (self *Link) upload(hub *Hub) {
	defer self.conn.CloseRead()
	rd := bufio.NewReaderSize(self.conn, 4096)
	for {
		buffer := make([]byte, 4096)
		n, err := rd.Read(buffer)
		if err != nil {
			self.setError(err)
			return
		}
		Debug("link(%d) read %d bytes:%s", self.id, n, string(buffer[:n]))
		hub.SendLinkData(self.id, buffer[:n])
	}
}

// read data from peer
func (self *Link) download(ch chan []byte) {
	defer self.conn.CloseWrite()

	for data := range ch {
		Debug("link(%d) write %d bytes:%s", self.id, len(data), string(data))
		_, err := self.conn.Write(data)
		if err != nil {
			self.setError(err)
			return
		}
	}
	// receive LINK_DESTROY, so close conn
	self.setError(errPeerClosed)
}

func (self *Link) Pump(hub *Hub, ch chan []byte) {
	go self.download(ch)
	self.upload(hub)

	if self.err != errPeerClosed {
		hub.Reset(self.id)
		hub.SendLinkDestory(self.id)
		Info("link(%d) closing: %v", self.id, self.err)
	}
}

func NewLink(id uint16, conn *net.TCPConn) *Link {
	conn.SetKeepAlive(true)
	conn.SetLinger(-1)
	link := new(Link)
	link.id = id
	link.conn = conn
	return link
}
