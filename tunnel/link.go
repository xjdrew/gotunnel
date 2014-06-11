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
func (self *Link) upload(coor *Coor) {
	rd := bufio.NewReaderSize(self.conn, 4096)
	for {
		buffer := make([]byte, 0xff)
		n, err := rd.Read(buffer)
		if err != nil {
			self.setError(err)
			return
		}
		Debug("link(%d) read %d bytes:%s", self.id, n, string(buffer[:n]))
		coor.SendLinkData(self.id, buffer[:n])
	}
}

// read data from peer
func (self *Link) download(ch chan []byte) {
	defer self.conn.Close()
	for data := range ch {
		if len(data) == 0 {
			break
		}

		c := 0
		for c < len(data) {
			n, err := self.conn.Write(data[c:])
			if err != nil {
				self.setError(err)
				return
			}
			c += n
		}
	}
	// receive LINK_DESTROY, so close conn
	self.setError(errPeerClosed)
}

func (self *Link) Pump(coor *Coor, ch chan []byte) {
	go self.download(ch)
	self.upload(coor)

	if self.err != errPeerClosed {
		coor.Reset(self.id)
		coor.SendLinkDestory(self.id)
		ch <- nil
		Info("link(%d) closing: %v", self.id, self.err)
	}
}

func NewLink(id uint16, conn *net.TCPConn) *Link {
	link := new(Link)
	link.id = id
	link.conn = conn
	return link
}
