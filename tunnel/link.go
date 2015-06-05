//
//   date  : 2014-06-04
//   author: xjdrew
//

package tunnel

import (
	"bufio"
	"errors"
	"sync"
)

var errPeerClosed = errors.New("errPeerClosed")

type Link struct {
	id    uint16
	conn  BiConn
	hub   *Hub
	rbuf  *LinkBuffer // 接收缓存
	sflag bool        // 对端是否可以收数据
	wg    sync.WaitGroup
}

// stop write data to remote
func (self *Link) resetSflag() bool {
	if self.sflag {
		self.sflag = false
		// close read
		if self.conn != nil {
			self.conn.CloseRead()
		}
		return true
	}
	return false
}

// stop recv data from remote
func (self *Link) resetRflag() bool {
	return self.rbuf.Close()
}

// peer link closed
func (self *Link) resetRSflag() bool {
	ok1 := self.resetSflag()
	ok2 := self.resetRflag()
	return ok1 || ok2
}

func (self *Link) SendCreate() {
	self.hub.Send(LINK_CREATE, self.id, nil)
}

func (self *Link) SendClose() {
	if self.resetRSflag() {
		self.hub.Send(LINK_CLOSE, self.id, nil)
	}
}

func (self *Link) putData(data []byte) bool {
	return self.rbuf.Put(data)
}

// read from link
func (self *Link) pumpIn() {
	defer self.wg.Done()
	defer self.conn.CloseRead()

	bufsize := PacketSize * 2
	rd := bufio.NewReaderSize(self.conn, bufsize)
	for {
		buffer := mpool.Get()
		n, err := rd.Read(buffer)
		if err != nil {
			if self.resetSflag() {
				self.hub.Send(LINK_CLOSE_SEND, self.id, nil)
			}
			mpool.Put(buffer)
			Debug("link(%d) read failed:%v", self.id, err)
			break
		}
		Trace("link(%d) read %d bytes:%s", self.id, n, string(buffer[:n]))

		if !self.sflag {
			// receive LINK_CLOSE_WRITE
			mpool.Put(buffer)
			break
		}
		if !self.hub.Send(LINK_DATA, self.id, buffer[:n]) {
			break
		}
	}
}

// write to link
func (self *Link) pumpOut() {
	defer self.wg.Done()
	defer self.conn.CloseWrite()

	for {
		data, ok := self.rbuf.Pop()
		if !ok {
			break
		}

		_, err := self.conn.Write(data)
		mpool.Put(data)

		if err != nil {
			if self.resetRflag() {
				self.hub.Send(LINK_CLOSE_RECV, self.id, nil)
			}
			Debug("link(%d) write failed:%v", self.id, err)
			break
		}
		Trace("link(%d) write %d bytes:%s", self.id, len(data), string(data))
	}
}

func (self *Link) Pump(conn BiConn) {
	self.conn = conn

	self.wg.Add(1)
	go self.pumpIn()

	self.wg.Add(1)
	go self.pumpOut()

	self.wg.Wait()
	Info("link(%d) closed", self.id)
}

func newLink(id uint16, hub *Hub) *Link {
	return &Link{
		id:    id,
		hub:   hub,
		rbuf:  NewLinkBuffer(16),
		sflag: true}
}
