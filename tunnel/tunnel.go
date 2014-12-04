//
//   date  : 2014-06-04
//   author: xjdrew
//

package tunnel

import (
	"bufio"
	"encoding/binary"
	"fmt"
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
	desc     string
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
func (self *Tunnel) PumpIn() (err error) {
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

		if header.Sz > options.PacketSize {
			Error("too long packet:%d", header.Sz)
			return
		}

		var data []byte
		if header.Sz > 0 {
			data = mpool.Get()[0:header.Sz]
			c := 0
			for c < int(header.Sz) {
				var n int
				n, err = rd.Read(data[c:])
				if err != nil {
					mpool.Put(data)
					Error("read tunnel failed:%s", err.Error())
					return
				}
				c += n
			}
		}

		self.outputCh <- &TunnelPayload{header.Linkid, data}
	}
	return
}

// write to tunnel
func (self *Tunnel) PumpOut() (err error) {
	var header struct {
		Linkid uint16
		Sz     uint16
	}

	wr := NewRC4Writer(self.conn, options.Rc4Key)
	for {
		payload := <-self.inputCh

		sz := len(payload.Data)
		if uint16(sz) > options.PacketSize {
			Panic("receive malformed payload, linkid:%d, sz:%d", payload.Linkid, sz)
			break
		}

		header.Linkid = payload.Linkid
		header.Sz = uint16(sz)
		err = binary.Write(wr, binary.LittleEndian, &header)
		if err != nil {
			Error("write tunnel failed:%s", err.Error())
			mpool.Put(payload.Data)
			return
		}

		_, err = wr.Write(payload.Data)
		mpool.Put(payload.Data)
		if err != nil {
			Error("write tunnel failed:%s", err.Error())
			return
		}
	}
	return
}

func (self *Tunnel) String() string {
	return self.desc
}

func newTunnel(conn *net.TCPConn) *Tunnel {
	conn.SetKeepAlive(true)
	conn.SetKeepAlivePeriod(time.Second * 60)
	conn.SetLinger(-1)
	desc := fmt.Sprintf("tunnel[%s <-> %s]", conn.LocalAddr(), conn.RemoteAddr())
	tunnel := new(Tunnel)
	tunnel.inputCh = make(chan *TunnelPayload, 65535)
	tunnel.outputCh = make(chan *TunnelPayload, 65535)
	tunnel.conn = conn
	tunnel.desc = desc
	return tunnel
}
