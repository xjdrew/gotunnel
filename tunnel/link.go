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
	id    uint16
	hub   *Hub
	conn  *net.TCPConn
	rbuf  *LinkBuffer // 接收缓存
	sflag bool        // 对端是否可以收数据
	qos   *Qos
	wg    sync.WaitGroup
}

// stop write data to remote
func (self *Link) resetSflag() bool {
	if self.sflag {
		self.sflag = false
		// close read
		self.conn.CloseRead()
		return true
	}
	return false
}

// stop recv data from remote
func (self *Link) resetRflag() bool {
	self.qos.Close()
	return self.rbuf.Close()
}

// peer link closed
func (self *Link) resetRSflag() bool {
	ok1 := self.resetSflag()
	ok2 := self.resetRflag()
	return ok1 || ok2
}

// set remote qos flag
func (self *Link) setRemoteQosFlag(flag bool) {
	self.qos.SetRemoteFlag(flag)
}

func (self *Link) SendCreate() {
	self.hub.Send(LINK_CREATE, self.id, nil)
}

func (self *Link) SendClose() {
	self.resetRSflag()
	self.hub.Send(LINK_CLOSE, self.id, nil)
}

func (self *Link) putData(data []byte) bool {
	ok := self.rbuf.Put(data)
	if ok {
		self.qos.SetWater(self.rbuf.Len())
	}
	return ok
}

func (self *Link) popData() (data []byte, ok bool) {
	data, ok = self.rbuf.Pop()
	if ok {
		self.qos.SetWater(self.rbuf.Len())
	}
	return
}

// read from link
func (self *Link) pumpIn() {
	defer self.wg.Done()
	defer self.conn.CloseRead()

	rd := bufio.NewReaderSize(self.conn, 4096)
	for {
		// qos balance
		self.qos.Balance()

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
		self.hub.Send(LINK_DATA, self.id, buffer[:n])
	}
}

// write to link
func (self *Link) pumpOut() {
	defer self.wg.Done()
	defer self.conn.CloseWrite()

	for {
		data, ok := self.popData()
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

func (self *Link) Pump(conn *net.TCPConn) {
	conn.SetKeepAlive(true)
	conn.SetLinger(-1)
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
		id:   id,
		hub:  hub,
		rbuf: NewLinkBuffer(16),
		qos: NewQos(options.RbufHw, options.RbufLw, func() {
			hub.Send(LINK_RECVBUF_HW, id, nil)
			Error("link(%d) enter high water", id)
		}, func() {
			hub.Send(LINK_RECVBUF_LW, id, nil)
			Error("link(%d) leave high water", id)
		}),
		sflag: true}
}
