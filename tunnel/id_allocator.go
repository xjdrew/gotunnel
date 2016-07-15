//
//   date  : 2015-08-31
//   author: xjdrew
//
package tunnel

type IdAllocator struct {
	freeList chan uint16
}

func (alloc *IdAllocator) Acquire() uint16 {
	return <-alloc.freeList
}

func (alloc *IdAllocator) Release(id uint16) {
	alloc.freeList <- id
}

func newAllocator() *IdAllocator {
	freeList := make(chan uint16, TunnelMaxId)
	var id uint16
	for id = 1; id != TunnelMaxId; id++ {
		freeList <- id
	}
	return &IdAllocator{freeList: freeList}
}
