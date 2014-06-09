package main

import (
	"bufio"
	"errors"
	"io"
	"net"
)

var ErrPeerClosed = errors.New("ErrPeerClosed")

type Link struct {
	id   uint16
	conn *net.TCPConn
	err  error
}

// write data to peer
func (self *Link) Upload(coor *Coor) {
	rd := bufio.NewReaderSize(self.conn, 255)
	t := 0
	for {
		buffer := make([]byte, 0xff)
		n, err := rd.Read(buffer)
		if err != nil {
			if self.err != nil {
				break
			}

			self.err = err
			if err == io.EOF {
				Debug("link(%d) upload finish", self.id)
			} else {
				Error("link(%d) upload failed:%v, err:%v, data len:%d", self.id, self.conn.LocalAddr(), err, t)
			}
			break
		}
		t += n
		Debug("link(%d) read %d bytes:%s", self.id, n, string(buffer[:n]))
		coor.SendLinkData(self.id, buffer[:n])
	}
}

// read data from peer
func (self *Link) Download(ch chan []byte) {
	for data := range ch {
		c := 0
		for c < len(data) {
			n, err := self.conn.Write(data[c:])
			if err != nil {
				if self.err == nil {
					self.err = err
					Error("link(%d) write failed:%v", self.id, err)
				}
				return
			}
			c += n
		}
	}

	// receive LINK_DESTROY, so close conn
	self.err = ErrPeerClosed
	self.conn.Close()
}

func (self *Link) Pump(coor *Coor, ch chan []byte) {
	go self.Download(ch)
	self.Upload(coor)

	if self.err != ErrPeerClosed {
		coor.Reset(self.id)
		coor.SendLinkDestory(self.id)
	}
}

func NewLink(id uint16, conn *net.TCPConn) *Link {
	link := new(Link)
	link.id = id
	link.conn = conn
	return link
}
