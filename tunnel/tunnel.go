//
//   date  : 2014-06-04
//   author: xjdrew
//

package tunnel

import (
	"bufio"
	"encoding/binary"
	"net"
	"time"
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

func (self *Tunnel) Close() {
	if self.conn != nil {
		self.conn.Close()
	}
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

// read from tunnel
func (self *Tunnel) PumpOut() (err error) {
	defer close(self.outputCh)

	var header struct {
		Linkid uint16
		Sz     uint16
	}

	rd := NewRC4Reader(bufio.NewReaderSize(self.conn, 8192), options.Rc4Key)
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

// write to tunnel
func (self *Tunnel) PumpUp() (err error) {
	var header struct {
		Linkid uint16
		Sz     uint16
	}

	wr := NewRC4Writer(self.conn, options.Rc4Key)
	for {
		payload := <-self.inputCh

		sz := len(payload.Data)
		if sz > 0xffff {
			Panic("receive malformed payload, linkid:%d, sz:%d", payload.Linkid, sz)
			break
		}

		header.Linkid = payload.Linkid
		header.Sz = uint16(sz)
		err = binary.Write(wr, binary.LittleEndian, &header)
		if err != nil {
			Error("write tunnel failed:%s", err.Error())
			return
		}

		_, err = wr.Write(payload.Data)
		if err != nil {
			Error("write tunnel failed:%s", err.Error())
			return
		}
	}
	return
}

func newTunnel(conn *net.TCPConn) *Tunnel {
	conn.SetKeepAlive(true)
	conn.SetKeepAlivePeriod(time.Second * 60)
	conn.SetLinger(-1)
	return &Tunnel{make(chan *TunnelPayload, 65535), make(chan *TunnelPayload, 65535), conn}
}
