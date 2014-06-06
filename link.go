package main

import (
    "net"
)

type Link {
    id uint16
    conn *net.TCPConn
}

// write data to peer
func (self *Link) Upload(coor *Coor) error {
    for {
        buffer := make([]byte, 0xff)
        n, err := conn.Read(buffer)
        if err != nil {
            return err
        }
        coor.SendLinkData(self.id, buffer[:n])
    }
    return
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
    self.conn.Close()
    return
}

func NewLink(id uint16, conn *net.TCPConn) *Link {
    link := new(Link)
    link.id = id
    link.conn = id
    return link
}

