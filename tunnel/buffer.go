//
//   date  : 2014-12-01
//   author: xjdrew
//

package tunnel

import "sync"

type Buffer struct {
	start  int
	end    int
	buf    [][]byte
	cond   *sync.Cond // buffer notify
	closed bool
}

func (b *Buffer) bufferLen() int {
	return (b.end + cap(b.buf) - b.start) % cap(b.buf)
}

func (b *Buffer) Len() int {
	b.cond.L.Lock()
	defer b.cond.L.Unlock()
	return b.bufferLen()
}

func (b *Buffer) Close() bool {
	b.cond.L.Lock()
	defer b.cond.L.Unlock()

	if b.closed {
		return false
	}

	b.closed = true
	b.cond.Broadcast()
	return true
}

func (b *Buffer) Put(data []byte) bool {
	b.cond.L.Lock()
	defer b.cond.L.Unlock()

	if b.closed {
		return false
	}

	// if there is only 1 free slot, we allocate more
	var old_cap = cap(b.buf)
	if (b.end+1)%old_cap == b.start {
		buf := make([][]byte, cap(b.buf)*2)
		if b.end > b.start {
			copy(buf, b.buf[b.start:b.end])
		} else if b.end < b.start {
			copy(buf, b.buf[b.start:old_cap])
			copy(buf[old_cap-b.start:], b.buf[0:b.end])
		}
		b.buf = buf
		b.start = 0
		b.end = old_cap - 1
	}

	b.buf[b.end] = data
	b.end = (b.end + 1) % cap(b.buf)
	b.cond.Signal()
	return true
}

func (b *Buffer) Pop() ([]byte, bool) {
	for {
		b.cond.L.Lock()
		if b.bufferLen() > 0 {
			data := b.buf[b.start]
			b.start = (b.start + 1) % cap(b.buf)
			b.cond.L.Unlock()
			return data, true
		}
		if b.closed {
			b.cond.L.Unlock()
			return nil, false
		}
		b.cond.Wait()
		b.cond.L.Unlock()
	}
}

func NewBuffer(sz int) *Buffer {
	var l sync.Mutex
	return &Buffer{
		buf:   make([][]byte, sz),
		start: 0,
		end:   0,
		cond:  sync.NewCond(&l),
	}
}
