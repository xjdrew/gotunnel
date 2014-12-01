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
	hub  *Hub
	conn *net.TCPConn

	rbuffer *LinkBuffer // 接收缓存
	sflag   bool        // 对端是否可以收数据

	wg sync.WaitGroup
}

// stop write data to remote
func (self *Link) resetSflag() bool {
	if self.sflag {
		self.sflag = false
		return true
	}
	return false
}

// stop recv data from remote
func (self *Link) resetRflag() bool {
	return self.rbuffer.Close()
}

func (self *Link) resetRSflag() bool {
	ok1 := self.resetSflag()
	ok2 := self.resetRflag()
	return ok1 || ok2
}

func (self *Link) putData(data []byte) bool {
	return self.rbuffer.Put(data)
}

func (self *Link) SendCreate() {
	self.hub.Send(LINK_CREATE, self.id, nil)
}

func (self *Link) SendClose() {
	self.resetRSflag()
	self.hub.Send(LINK_CLOSE, self.id, nil)
}

// write data to peer
func (self *Link) send() {
	linkid := self.id

	defer self.wg.Done()
	defer self.conn.CloseRead()

	rd := bufio.NewReaderSize(self.conn, 4096)
	for {
		buffer := make([]byte, 4096)
		n, err := rd.Read(buffer)
		if err != nil {
			if self.resetSflag() {
				self.hub.Send(LINK_CLOSE_SEND, self.id, nil)
			}
			Debug("link(%d) read failed:%v", linkid, err)
			break
		}
		Debug("link(%d) read %d bytes:%s", linkid, n, string(buffer[:n]))

		if !self.sflag {
			// receive LINK_CLOSE_WRITE
			break
		}
		self.hub.Send(LINK_DATA, self.id, buffer[:n])
	}
}

// read data from peer
func (self *Link) recv() {
	defer self.wg.Done()
	defer self.conn.CloseWrite()

	for {
		data, ok := self.rbuffer.Pop()
		if !ok {
			break
		}
		_, err := self.conn.Write(data)
		if err != nil {
			if self.resetRflag() {
				self.hub.Send(LINK_CLOSE_RECV, self.id, nil)
			}
			Debug("link(%d) write failed:%v", self.id, err)
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
	go self.recv()

	self.wg.Add(1)
	go self.send()

	self.wg.Wait()
	Info("link(%d) closed", self.id)
}

func newLink(id uint16, hub *Hub) *Link {
	link := new(Link)
	link.id = id
	link.hub = hub

	link.rbuffer = NewLinkBuffer(16)
	link.sflag = true
	return link
}
