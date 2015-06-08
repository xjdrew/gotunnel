//
//   date  : 2014-06-04
//   author: xjdrew
//

package tunnel

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

var Timeout int64 // tunnel read/write timeout

type Payload struct {
	linkid uint16
	data   []byte
}

type Tunnel struct {
	conn   *net.TCPConn  // low level conn
	writer *bufio.Writer // writer
	reader *bufio.Reader // reader
	wch    chan Payload  // write data chan
	closed chan struct{} // connection closed
	once   sync.Once
	desc   string // description
}

func (t *Tunnel) shutdown() {
	t.conn.Close()
	close(t.closed)
}

func (t *Tunnel) Close() {
	t.once.Do(t.shutdown)
}

func (t *Tunnel) write(payload Payload) error {
	defer mpool.Put(payload.data)

	if err := binary.Write(t.writer, binary.LittleEndian, payload.linkid); err != nil {
		return err
	}
	if err := binary.Write(t.writer, binary.LittleEndian, uint16(len(payload.data))); err != nil {
		return err
	}
	if _, err := t.writer.Write(payload.data); err != nil {
		return err
	}
	if err := t.writer.Flush(); err != nil {
		return err
	}
	return nil
}

func (t *Tunnel) pump() {
	for {
		select {
		case payload := <-t.wch:
			if err := t.write(payload); err != nil {
				t.once.Do(t.shutdown)
				Error("%s write failed:%v", t.desc, err)
				return
			}
		case <-t.closed:
			Error("%s closed", t.desc)
			return
		}
	}
}

func (t *Tunnel) Write(payload Payload) bool {
	select {
	case t.wch <- payload:
		return true
	case <-t.closed:
		return false
	}
}

func (t *Tunnel) Read() (Payload, error) {
	var payload Payload
	var linkid, sz uint16

	// disable timeout when read packet head
	t.conn.SetReadDeadline(time.Time{})
	if err := binary.Read(t.reader, binary.LittleEndian, &linkid); err != nil {
		return payload, err
	}

	if err := binary.Read(t.reader, binary.LittleEndian, &sz); err != nil {
		return payload, err
	}

	data := mpool.Get()[0:sz]
	// timeout if can't read a packet in 10 seconds
	if Timeout > 0 {
		t.conn.SetReadDeadline(time.Now().Add(time.Duration(Timeout) * time.Second))
	}
	if _, err := io.ReadFull(t.reader, data); err != nil {
		return payload, err
	}
	payload.linkid = linkid
	payload.data = data
	return payload, nil
}

func (self *Tunnel) String() string {
	return self.desc
}

func newTunnel(conn *net.TCPConn, rc4key []byte) *Tunnel {
	desc := fmt.Sprintf("tunnel[%s <-> %s]", conn.LocalAddr(), conn.RemoteAddr())
	conn.SetKeepAlive(true)
	conn.SetKeepAlivePeriod(time.Second * 60)
	bufsize := int(PacketSize) * 2
	tunnel := &Tunnel{
		writer: bufio.NewWriterSize(NewRC4Writer(conn, rc4key), bufsize),
		reader: bufio.NewReaderSize(NewRC4Reader(conn, rc4key), bufsize),
		wch:    make(chan Payload),
		closed: make(chan struct{}),
		conn:   conn,
		desc:   desc,
	}

	go tunnel.pump()
	return tunnel
}
