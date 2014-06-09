package main

import (
	"bufio"
	"net"
    "io"
)

type Link struct {
	id         uint16
	conn       *net.TCPConn
	peerClosed bool
}

// write data to peer
func (self *Link) Upload(coor *Coor) (err error) {
	rd := bufio.NewReaderSize(self.conn, 4096)
    t := 0
	for {
		buffer := make([]byte, 0xff)
		var n int
		n, err = rd.Read(buffer)
		if err != nil {
			if self.peerClosed {
                Info("link(%d) peer closed", self.id)
				return nil
			}

            if err != io.EOF {
                Error("link(%d) upload failed:%v, err:%v, data len:%d", self.id, self.conn.LocalAddr(), err, t)
            } else {
                Info("link(%d) upload finish", self.id)
            }
			coor.SendLinkDestory(self.id)
			return err
		}
        t += n
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
                Error("link(%d) write failed:%v", self.id, err)
				return err
			}
			c += n
		}
	}

	// receive LINK_DESTROY, so close conn
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

