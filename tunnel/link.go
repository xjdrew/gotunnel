//
//   date  : 2014-06-04
//   author: xjdrew
//

package tunnel

import (
	"errors"
	"net"
	"sync"
	"time"
)

var errPeerClosed = errors.New("errPeerClosed")

type link struct {
	id   uint16
	conn *net.TCPConn
	wbuf *Buffer // write buffer

	lock sync.Mutex // protects below fields
	rerr error      // if read closed, error to give reads
}

// set rerr
func (l *link) setRerr(err error) bool {
	l.lock.Lock()
	defer l.lock.Unlock()

	if l.rerr != nil {
		return false
	}

	l.rerr = err
	if l.conn != nil {
		l.conn.CloseRead()
	}
	return true
}

// stop read data from link
func (l *link) rclose() bool {
	return l.setRerr(errPeerClosed)
}

// stop write data into link
func (l *link) wclose() bool {
	return l.wbuf.Close()
}

// close link
func (l *link) aclose() {
	l.rclose()
	l.wclose()
}

// read data from link
func (l *link) read() ([]byte, error) {
	if l.rerr != nil {
		return nil, l.rerr
	}
	b := mpool.Get()
	n, err := l.conn.Read(b)
	if err != nil {
		l.setRerr(err)
		return nil, l.rerr
	}
	if l.rerr != nil {
		return nil, l.rerr
	}
	return b[:n], nil
}

// write data into link
func (l *link) write(b []byte) bool {
	return l.wbuf.Put(b)
}

// inject data low level connection
func (l *link) _write() error {
	defer l.conn.CloseWrite()

	for {
		data, ok := l.wbuf.Pop()
		if !ok {
			return errPeerClosed
		}

		_, err := l.conn.Write(data)
		mpool.Put(data)
		if err != nil {
			return err
		}
	}
}

// set low level connection
func (l *link) setConn(conn *net.TCPConn) {
	if l.conn != nil {
		Panic("link(%d) repeated set conn", l.id)
	}
	l.conn = conn
}

// hub function
func (h *Hub) getLink(id uint16) *link {
	h.ll.RLock()
	defer h.ll.RUnlock()
	return h.links[id]
}

func (h *Hub) deleteLink(id uint16) {
	Info("link(%d) delete", id)
	h.ll.Lock()
	defer h.ll.Unlock()
	delete(h.links, id)
}

func (h *Hub) createLink(id uint16) *link {
	Info("link(%d) new link", id)
	h.ll.Lock()
	defer h.ll.Unlock()
	if _, ok := h.links[id]; ok {
		Error("link(%d) repeated", id)
		return nil
	}
	l := &link{
		id:   id,
		wbuf: NewBuffer(16),
	}
	h.links[id] = l
	return l
}

func (h *Hub) startLink(l *link, conn *net.TCPConn) {
	conn.SetKeepAlive(true)
	conn.SetKeepAlivePeriod(time.Second * 60)
	l.setConn(conn)

	Info("link(%d) start: %v", l.id, conn.RemoteAddr())
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			data, err := l.read()
			if err != nil {
				if err != errPeerClosed {
					h.SendCmd(l.id, LINK_CLOSE_SEND)
				}
				break
			}

			h.Send(l.id, data)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := l._write()
		if err != errPeerClosed {
			h.SendCmd(l.id, LINK_CLOSE_RECV)
		}
	}()
	wg.Wait()
	Info("link(%d) close", l.id)
}
