//
//   date  : 2014-06-04
//   author: xjdrew
//

package tunnel

import (
	"bufio"
	"encoding/binary"
	"net"
)

type TunnelPayload struct {
	Linkid uint16
	Data   []byte
}

type Tunnel struct {
	inputCh  chan *TunnelPayload
	outputCh chan *TunnelPayload
	conn     *net.TCPConn
}

func (self *Tunnel) Put(payload *TunnelPayload) {
	self.inputCh <- payload
}

func (self *Tunnel) Pop() *TunnelPayload {
	payload, ok := <-self.outputCh
	if !ok {
		return nil
	}
	return payload
}

// read
func (self *Tunnel) PumpOut() (err error) {
	defer close(self.outputCh)

	var header struct {
		Linkid uint16
		Sz     uint8
	}

	rd := NewRC4Reader(bufio.NewReaderSize(self.conn, 4096), options.Rc4Key)
	for {
		err = binary.Read(rd, binary.LittleEndian, &header)
		if err != nil {
			Error("read tunnel failed:%s", err.Error())
			return
		}

		var data []byte

		// if header.Sz == 0, it's ok too
		data = make([]byte, header.Sz)
		c := 0
		for c < int(header.Sz) {
			var n int
			n, err = rd.Read(data[c:])
			if err != nil {
				Error("read tunnel failed:%s", err.Error())
				return
			}
			c += n
		}

		self.outputCh <- &TunnelPayload{header.Linkid, data}
	}
	return
}

// write
func (self *Tunnel) PumpUp() (err error) {
	var header struct {
		Linkid uint16
		Sz     uint8
	}

	wr := NewRC4Writer(self.conn, options.Rc4Key)
	for {
		payload := <-self.inputCh

		sz := len(payload.Data)
		if sz > 0xff {
			Panic("receive malformed payload, linkid:%d, sz:%d", payload.Linkid, sz)
			break
		}

		header.Linkid = payload.Linkid
		header.Sz = uint8(sz)
		err = binary.Write(wr, binary.LittleEndian, &header)
		if err != nil {
			Error("write tunnel failed:%s", err.Error())
			return
		}

		c := 0
		for c < sz {
			var n int
			n, err = wr.Write(payload.Data[c:])
			if err != nil {
				Error("write tunnel failed:%s", err.Error())
				return
			}
			c += n
		}
	}
	return
}

func NewTunnel(conn *net.TCPConn) *Tunnel {
	return &Tunnel{make(chan *TunnelPayload, 65535), make(chan *TunnelPayload, 65535), conn}
}
