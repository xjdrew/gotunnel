//
//   date  : 2014-06-04
//   author: xjdrew
//

package tunnel

import (
	"bufio"
	"errors"
	"net"
	"sync"
)

var errPeerClosed = errors.New("errPeerClosed")

type Link struct {
	id     uint16
	hub    *Hub
	conn   *net.TCPConn
	reader func() ([]byte, bool)
	wg     sync.WaitGroup
}

// write data to peer
func (self *Link) upload() {
	linkid := self.id

	defer self.wg.Done()
	defer self.conn.CloseRead()

	rd := bufio.NewReaderSize(self.conn, 4096)
	for {
		buffer := make([]byte, 4096)
		n, err := rd.Read(buffer)
		if err != nil {
			Debug("link(%d) read failed:%v", linkid, err)
			self.hub.SendLinkCloseRead(linkid)
			break
		}
		Debug("link(%d) read %d bytes:%s", linkid, n, string(buffer[:n]))
		if !self.hub.SendLinkData(linkid, buffer[:n]) {
			break
		}
	}
}

// read data from peer
func (self *Link) download() {
	linkid := self.id

	defer self.wg.Done()
	defer self.conn.CloseWrite()

	for {
		data, ok := self.reader()
		if !ok {
			break
		}
		_, err := self.conn.Write(data)
		if err != nil {
			Debug("link(%d) write failed:%v", linkid, err)
			self.hub.SendLinkCloseWrite(self.id)
			break
		}
		Debug("link(%d) write %d bytes:%s", self.id, len(data), string(data))
	}
}

func (self *Link) Pump(conn *net.TCPConn) {
	conn.SetKeepAlive(true)
	conn.SetLinger(-1)
	self.conn = conn

	self.wg.Add(1)
	go self.download()

	self.wg.Add(1)
	go self.upload()

	self.wg.Wait()
	Info("link(%d) closed", self.id)
}

func newLink(id uint16, hub *Hub, reader func() ([]byte, bool)) *Link {
	link := new(Link)
	link.id = id
	link.hub = hub
	link.reader = reader
	return link
}
