//
//   date  : 2014-12-04
//   author: xjdrew
//

package tunnel

import (
	"sync"
)

type MPool struct {
	*sync.Pool
	sz int
}

func (p *MPool) Get() []byte {
	return p.Pool.Get().([]byte)
}

func (p *MPool) Put(x []byte) {
	if cap(x) == p.sz {
		p.Pool.Put(x[0:p.sz])
	}
}

func NewMPool(sz int) *MPool {
	p := &MPool{sz: sz}
	p.Pool = &sync.Pool{
		New: func() interface{} {
			return make([]byte, p.sz)
		},
	}
	return p
}
