//
//   date  : 2014-08-27
//   author: xjdrew
//

package tunnel

import (
	"crypto/rc4"
	"net"
)

type Conn struct {
	net.Conn
	enc *rc4.Cipher
	dec *rc4.Cipher
}

func (conn *Conn) SetCipherKey(key []byte) {
	conn.enc, _ = rc4.NewCipher(key)
	conn.dec, _ = rc4.NewCipher(key)
}

func (conn *Conn) Read(b []byte) (int, error) {
	n, err := conn.Conn.Read(b)
	if n > 0 && conn.dec != nil {
		conn.dec.XORKeyStream(b[:n], b[:n])
	}
	return n, err
}

func (conn *Conn) Write(b []byte) (int, error) {
	if conn.enc != nil {
		conn.enc.XORKeyStream(b, b)
	}
	return conn.Conn.Write(b)
}
