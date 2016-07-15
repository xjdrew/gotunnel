//
//   date  : 2014-08-27
//   author: xjdrew
//

package tunnel

import (
	"bufio"
	"crypto/rc4"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
)

var errTooLarge = fmt.Errorf("tunnel.Read: packet too large")

type TunnelConn struct {
	net.Conn
	reader *bufio.Reader
	writer *bufio.Writer
	enc    *rc4.Cipher
	dec    *rc4.Cipher
}

func (conn *TunnelConn) SetCipherKey(key []byte) {
	conn.enc, _ = rc4.NewCipher(key)
	conn.dec, _ = rc4.NewCipher(key)
}

func (conn *TunnelConn) Read(b []byte) (int, error) {
	n, err := conn.reader.Read(b)
	if n > 0 && conn.dec != nil {
		conn.dec.XORKeyStream(b[:n], b[:n])
	}
	return n, err
}

func (conn *TunnelConn) Write(b []byte) (int, error) {
	if conn.enc != nil {
		conn.enc.XORKeyStream(b, b)
	}
	return conn.writer.Write(b)
}

func (conn *TunnelConn) Flush() error {
	return conn.writer.Flush()
}

// tunnel packet header
// a tunnel packet consists of a header and a body
// Len is the length of subsequent packet body
type header struct {
	Linkid uint16
	Len    uint16
}

type Tunnel struct {
	*TunnelConn

	wlock sync.Mutex // protect concurrent write
	werr  error      // write error
}

// can write concurrently
func (tun *Tunnel) WritePacket(linkid uint16, data []byte) (err error) {
	defer mpool.Put(data)

	tun.wlock.Lock()
	defer tun.wlock.Unlock()

	if tun.werr != nil {
		return tun.werr
	}

	if err = binary.Write(tun, binary.LittleEndian, header{linkid, uint16(len(data))}); err != nil {
		tun.werr = err
		tun.Close()
		return err
	}

	if _, err = tun.Write(data); err != nil {
		tun.werr = err
		tun.Close()
		return err
	}

	if err = tun.Flush(); err != nil {
		tun.werr = err
		tun.Close()
		return err
	}
	return
}

// can't read concurrently
func (tun *Tunnel) ReadPacket() (linkid uint16, data []byte, err error) {
	var h header

	if err = binary.Read(tun, binary.LittleEndian, &h); err != nil {
		return
	}

	if h.Len > TunnelPacketSize {
		err = errTooLarge
		return
	}

	data = mpool.Get()[0:h.Len]
	if _, err = io.ReadFull(tun, data); err != nil {
		return
	}
	linkid = h.Linkid
	return
}

func (tun Tunnel) String() string {
	return fmt.Sprintf("tunnel[%s -> %s]", tun.Conn.LocalAddr(), tun.Conn.RemoteAddr())
}

func newTunnel(conn net.Conn) *Tunnel {
	var tun Tunnel
	tun.TunnelConn = &TunnelConn{conn, bufio.NewReaderSize(conn, TunnelPacketSize*2), bufio.NewWriterSize(conn, TunnelPacketSize*2), nil, nil}
	return &tun
}
