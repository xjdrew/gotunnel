//
//   date  : 2014-12-01
//   author: xjdrew
//

package tunnel

import "sync"

type LinkBuffer struct {
	ch   chan []byte // wake up pop
	off  int         // pop at &buf[off], put at &buf[len(buf)]
	buf  [][]byte
	lock sync.Mutex
}

func (b *LinkBuffer) Len() int {
	return len(b.buf) - b.off
}

func (b *LinkBuffer) Close() bool {
	b.lock.Lock()
	defer b.lock.Unlock()

	if b.ch == nil {
		return false
	}

	close(b.ch)
	b.ch = nil
	return true
}

func (b *LinkBuffer) Put(data []byte) bool {
	b.lock.Lock()
	defer b.lock.Unlock()

	if b.ch == nil {
		return false
	}
	m := b.Len()
	if len(b.buf) == cap(b.buf) {
		var buf [][]byte
		if b.off > 0 {
			copy(b.buf[:], b.buf[b.off:])
			buf = b.buf[:m]
		} else {
			// no enough space
			buf = make([][]byte, cap(b.buf)*2)
			copy(buf, b.buf[:])
		}
		b.buf = buf
		b.off = 0
	}

	n := b.off + m
	b.buf = b.buf[0 : n+1]
	b.buf[n] = data

	select {
	case b.ch <- b.buf[b.off]:
		b.off += 1
	default:
	}
	return true
}

func (b *LinkBuffer) Pop() (data []byte, ok bool) {
	b.lock.Lock()
	ok = true
	if b.off < len(b.buf) {
		data = b.buf[b.off]
		b.off += 1
		b.lock.Unlock()
		return
	}
	if b.ch == nil {
		ok = false
		return
	}
	b.lock.Unlock()

	// waiting for new data
	data, ok = <-b.ch
	return
}

func NewLinkBuffer(sz int) *LinkBuffer {
	return &LinkBuffer{
		ch:  make(chan []byte),
		buf: make([][]byte, sz)[0:0]}
}
