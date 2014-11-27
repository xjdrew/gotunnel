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
	id   uint16
	conn *net.TCPConn
	wg   sync.WaitGroup
}

// write data to peer
func (self *Link) upload(hub *Hub) {
	linkid := self.id

	defer self.wg.Done()
	defer self.conn.CloseRead()

	rd := bufio.NewReaderSize(self.conn, 4096)
	for {
		buffer := make([]byte, 4096)
		n, err := rd.Read(buffer)
		if err != nil {
			Debug("link(%d) read failed:%v", linkid, err)
			hub.SendLinkCloseRead(linkid)
			break
		}
		Debug("link(%d) read %d bytes:%s", linkid, n, string(buffer[:n]))
		if !hub.SendLinkData(linkid, buffer[:n]) {
			break
		}
	}
}

// read data from peer
func (self *Link) download(hub *Hub) {
	linkid := self.id

	defer self.wg.Done()
	defer self.conn.CloseWrite()

	reader := hub.RecvLinkData(linkid)
	for {
		data, ok := reader()
		if !ok {
			break
		}
		_, err := self.conn.Write(data)
		if err != nil {
			Debug("link(%d) write failed:%v", linkid, err)
			hub.SendLinkCloseWrite(self.id)
			break
		}
		Debug("link(%d) write %d bytes:%s", self.id, len(data), string(data))
	}
}

func (self *Link) Pump(hub *Hub) {
	self.wg.Add(1)
	go self.download(hub)

	self.wg.Add(1)
	go self.upload(hub)

	self.wg.Wait()
	Info("link(%d) closed", self.id)
}

func NewLink(id uint16, conn *net.TCPConn) *Link {
	conn.SetKeepAlive(true)
	conn.SetLinger(-1)
	link := new(Link)
	link.id = id
	link.conn = conn
	return link
}
