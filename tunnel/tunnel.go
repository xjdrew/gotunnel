//
//   date  : 2014-06-04
//   author: xjdrew
//

package tunnel

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"
)

type TunnelPayload struct {
	Linkid uint16
	Data   []byte
}

type TunnelHeader struct {
	Linkid uint16
	Sz     uint16
}

type Tunnel struct {
	wlock  *sync.Mutex  // write lock
	writer *RC4Writer   // writer
	rlock  *sync.Mutex  // read lock
	reader *RC4Reader   // reader
	conn   *net.TCPConn // low level conn
	desc   string       // description
}

func (t *Tunnel) Close() {
	t.conn.Close()
}

func (t *Tunnel) Write(payload TunnelPayload) (err error) {
	defer mpool.Put(payload.Data)

	var header TunnelHeader
	header.Linkid = payload.Linkid
	header.Sz = uint16(len(payload.Data))

	t.wlock.Lock()
	defer t.wlock.Unlock()
	err = binary.Write(t.writer, binary.LittleEndian, &header)
	if err != nil {
		return
	}
	_, err = t.writer.Write(payload.Data)
	return err
}

func (t *Tunnel) Read() (payload *TunnelPayload, err error) {
	t.rlock.Lock()
	defer t.rlock.Unlock()

	var header TunnelHeader
	err = binary.Read(t.reader, binary.LittleEndian, &header)
	if err != nil {
		return
	}

	if header.Sz > options.PacketSize {
		err = errors.New("malformed packet, too long")
		return
	}

	payload = &TunnelPayload{}
	payload.Linkid = header.Linkid
	data := mpool.Get()[0:header.Sz]
	c := 0
	for c < int(header.Sz) {
		var n int
		n, err = t.reader.Read(data[c:])
		if err != nil {
			mpool.Put(data)
			return
		}
		c += n
	}
	payload.Data = data
	return
}

func (self *Tunnel) String() string {
	return fmt.Sprintf("%s", self.desc)
}

func newTunnel(conn *net.TCPConn, rc4key []byte) *Tunnel {
	desc := fmt.Sprintf("tunnel[%s <-> %s]", conn.LocalAddr(), conn.RemoteAddr())
	if err := conn.SetReadBuffer(options.TunnelReadBuffer); err != nil {
		Error("%s set read buffer failed:%s", desc, err.Error())
	}
	if err := conn.SetWriteBuffer(options.TunnelWriteBuffer); err != nil {
		Error("%s set read buffer failed:%s", desc, err.Error())
	}
	conn.SetKeepAlive(true)
	conn.SetKeepAlivePeriod(time.Second * 60)
	return &Tunnel{
		wlock:  new(sync.Mutex),
		writer: NewRC4Writer(conn, rc4key),
		rlock:  new(sync.Mutex),
		reader: NewRC4Reader(bufio.NewReaderSize(conn, 8192), rc4key),
		conn:   conn,
		desc:   desc,
	}
}
