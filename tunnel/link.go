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
	"time"
)

var errPeerClosed = errors.New("errPeerClosed")

type Link struct {
	id    uint16
	conn  *net.TCPConn
	hub   *Hub
	rbuf  *LinkBuffer // 接收缓存
	sflag bool        // 对端是否可以收数据
	wg    sync.WaitGroup
}

// stop write data to remote
func (link *Link) resetSflag() bool {
	if link.sflag {
		link.sflag = false
		// close read
		if link.conn != nil {
			link.conn.CloseRead()
		}
		return true
	}
	return false
}

// stop recv data from remote
func (link *Link) resetRflag() bool {
	return link.rbuf.Close()
}

// peer link closed
func (link *Link) resetRSflag() bool {
	ok1 := link.resetSflag()
	ok2 := link.resetRflag()
	return ok1 || ok2
}

func (link *Link) SendCreate() {
	link.hub.Send(LINK_CREATE, link.id, nil)
}

func (link *Link) SendClose() {
	if link.resetRSflag() {
		link.hub.Send(LINK_CLOSE, link.id, nil)
	}
}

func (link *Link) putData(data []byte) bool {
	return link.rbuf.Put(data)
}

// read from link
func (link *Link) pumpIn() {
	defer link.wg.Done()
	defer link.conn.CloseRead()

	bufsize := PacketSize * 2
	rd := bufio.NewReaderSize(link.conn, bufsize)
	for {
		buffer := mpool.Get()
		n, err := rd.Read(buffer)
		if err != nil {
			if link.resetSflag() {
				link.hub.Send(LINK_CLOSE_SEND, link.id, nil)
			}
			mpool.Put(buffer)
			Debug("link(%d) read failed:%v", link.id, err)
			break
		}
		Trace("link(%d) read %d bytes:%s", link.id, n, string(buffer[:n]))

		if !link.sflag {
			// receive LINK_CLOSE_WRITE
			mpool.Put(buffer)
			break
		}
		if !link.hub.Send(LINK_DATA, link.id, buffer[:n]) {
			break
		}
	}
}

// write to link
func (link *Link) pumpOut() {
	defer link.wg.Done()
	defer link.conn.CloseWrite()

	for {
		data, ok := link.rbuf.Pop()
		if !ok {
			break
		}

		_, err := link.conn.Write(data)
		mpool.Put(data)

		if err != nil {
			if link.resetRflag() {
				link.hub.Send(LINK_CLOSE_RECV, link.id, nil)
			}
			Debug("link(%d) write failed:%v", link.id, err)
			break
		}
		Trace("link(%d) write %d bytes:%s", link.id, len(data), string(data))
	}
}

func (link *Link) Pump(conn *net.TCPConn) {
	conn.SetKeepAlive(true)
	conn.SetKeepAlivePeriod(time.Second * 60)
	link.conn = conn

	link.wg.Add(1)
	go link.pumpIn()

	link.wg.Add(1)
	go link.pumpOut()

	link.wg.Wait()
	Info("link(%d) closed", link.id)
	link.hub.deleteLink(link.id)
}

func newLink(id uint16, hub *Hub) *Link {
	link := &Link{
		id:    id,
		hub:   hub,
		rbuf:  NewLinkBuffer(16),
		sflag: true,
	}
	if hub.setLink(id, link) {
		return link
	}
	return nil
}
