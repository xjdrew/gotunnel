//
//   date  : 2014-06-04
//   author: xjdrew
//

package tunnel

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

var Timeout int64 // tunnel read/write timeout

type header struct {
	Linkid uint16
	Len    uint16
}

type Tunnel struct {
	*Conn
	wlock sync.Mutex // protect concurrent write
}

// can write concurrently
func (tun *Tunnel) Write(linkid uint16, data []byte) (err error) {
	defer mpool.Put(data)

	tun.wlock.Lock()
	defer tun.wlock.Unlock()

	if err = binary.Write(tun.Conn, binary.LittleEndian, header{linkid, uint16(len(data))}); err != nil {
		return err
	}
	if _, err = tun.Conn.Write(data); err != nil {
		return err
	}
	return
}

// can't read concurrently
func (tun *Tunnel) Read() (linkid uint16, data []byte, err error) {
	var h header

	// disable timeout when read packet head
	tun.SetReadDeadline(time.Time{})
	if err = binary.Read(tun.Conn, binary.LittleEndian, &h); err != nil {
		return
	}

	data = mpool.Get()[0:h.Len]
	// timeout if can't read a packet in time
	if Timeout > 0 {
		tun.SetReadDeadline(time.Now().Add(time.Duration(Timeout) * time.Second))
	}
	if _, err = io.ReadFull(tun.Conn, data); err != nil {
		return
	}
	linkid = h.Linkid
	return
}

func (tun *Tunnel) String() string {
	return fmt.Sprintf("tunnel[%s -> %s]", tun.Conn.LocalAddr(), tun.Conn.RemoteAddr())
}

func newTunnel(conn net.Conn, key []byte) *Tunnel {
	var tun Tunnel
	tun.Conn = &Conn{conn, nil, nil}
	tun.Conn.SetCipherKey(key)
	return &tun
}
