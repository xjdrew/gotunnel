//
//   date  : 2014-12-04
//   author: xjdrew
//

package tunnel

import (
	"sync"
	"sync/atomic"
)

type MPool struct {
	pool    *sync.Pool
	alloced int32
	used    int32
	sz      int
}

func (p *MPool) Get() []byte {
	atomic.AddInt32(&p.used, 1)
	return p.pool.Get().([]byte)
}

func (p *MPool) Put(x []byte) {
	if cap(x) == p.sz {
		atomic.AddInt32(&p.used, -1)
		p.pool.Put(x[0:p.sz])
	}
}

func (p *MPool) Alloced() int32 {
	return p.alloced
}

func (p *MPool) Used() int32 {
	return p.used
}

func NewMPool(sz int) *MPool {
	p := &MPool{sz: sz}
	pool := new(sync.Pool)
	pool.New = func() interface{} {
		buf := make([]byte, p.sz)
		atomic.AddInt32(&p.alloced, 1)
		return buf
	}
	p.pool = pool
	return p
}
