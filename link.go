package main

import (
	"net"
)

type Link struct {
	id         uint16
	conn       *net.TCPConn
	peerClosed bool
}

// write data to peer
func (self *Link) Upload(coor *Coor) error {
	for {
		buffer := make([]byte, 0xff)
		n, err := self.conn.Read(buffer)
		if err != nil {
			if self.peerClosed {
				return nil
			}
			coor.SendLinkDestory(self.id)
			return err
		}
		coor.SendLinkData(self.id, buffer[:n])
	}
	return nil
}

// read data from peer
func (self *Link) Download(ch chan []byte) error {
	for data := range ch {
		c := 0
		for c < len(data) {
			n, err := self.conn.Write(data[c:])
			if err != nil {
				return err
			}
			c += n
		}
	}

	// receive LINK_DESTORY, so close conn
	self.peerClosed = true
	self.conn.Close()
	return nil
}

func (self *Link) Pump(coor *Coor, ch chan []byte) error {
	go self.Download(ch)
	return self.Upload(coor)
}

func NewLink(id uint16, conn *net.TCPConn) *Link {
	link := new(Link)
	link.id = id
	link.conn = conn
	link.peerClosed = false
	return link
}
