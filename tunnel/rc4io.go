//
//   date  : 2014-06-11
//   author: xjdrew
//

package tunnel

import (
	"crypto/rc4"
	"io"
)

type RC4Reader struct {
	rd io.Reader
	c  *rc4.Cipher
}

func (r *RC4Reader) Read(p []byte) (n int, err error) {
	n, err = r.rd.Read(p)
	if n > 0 && r.c != nil {
		r.c.XORKeyStream(p[:n], p[:n])
	}
	return
}

func NewRC4Reader(rd io.Reader, key []byte) *RC4Reader {
	c, _ := rc4.NewCipher(key)
	return &RC4Reader{rd, c}
}

type RC4Writer struct {
	wr io.Writer
	c  *rc4.Cipher
}

func (w *RC4Writer) Write(p []byte) (int, error) {
	if w.c != nil {
		cp := make([]byte, len(p))
		w.c.XORKeyStream(cp, p)
		return w.wr.Write(cp)
	}
	return w.wr.Write(p)
}

func NewRC4Writer(wr io.Writer, key []byte) *RC4Writer {
	c, _ := rc4.NewCipher(key)
	return &RC4Writer{wr, c}
}
