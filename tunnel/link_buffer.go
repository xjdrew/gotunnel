//
//   date  : 2014-12-01
//   author: xjdrew
//

package tunnel

import "sync"

type LinkBuffer struct {
	ch    chan []byte // wake up pop
	start int
	end   int
	buf   [][]byte
	lock  sync.Mutex // protect ch && buf
}

func (b *LinkBuffer) Len() int {
	if b.start == b.end {
		return 0
	} else if b.end > b.start {
		return b.end - b.start
	} else {
		return cap(b.buf) - b.start + b.end
	}
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
	select {
	case b.ch <- b.buf[b.start]:
		b.start = (b.start + 1) % cap(b.buf)
	default:
	}
	return true
}

func (b *LinkBuffer) Pop() (data []byte, ok bool) {
	b.lock.Lock()
	ok = true
	if b.Len() > 0 {
		data = b.buf[b.start]
		b.start = (b.start + 1) % cap(b.buf)
		b.lock.Unlock()
		return
	}
	if b.ch == nil {
		ok = false
		b.lock.Unlock()
		return
	}
	b.lock.Unlock()

	// waiting for new data
	data, ok = <-b.ch
	return
}

func NewLinkBuffer(sz int) *LinkBuffer {
	return &LinkBuffer{
		ch:    make(chan []byte),
		buf:   make([][]byte, sz),
		start: 0,
		end:   0,
	}
}
