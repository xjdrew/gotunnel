//
//   date  : 2014-06-04
//   author: xjdrew
//

package tunnel

import (
	"errors"
	"sync"
)

var IndexError = errors.New("index out of range or no data")
var ConflictError = errors.New("linkid conflict")

type LinkSet struct {
	capacity   uint16
	chs        []chan []byte
	freeLinkid chan uint16
	rw         sync.RWMutex
}

func (self *LinkSet) isValidLinkid(linkid uint16) bool {
	if 0 < linkid && linkid < self.capacity {
		return true
	}
	return false
}

func (self *LinkSet) AcquireId() uint16 {
	var linkid uint16 = 0
	select {
	case linkid = <-self.freeLinkid:
	default:
		Error("allocate linkid failed")
	}
	return linkid
}

func (self *LinkSet) ReleaseId(linkid uint16) (err error) {
	if !self.isValidLinkid(linkid) {
		err = IndexError
		return
	}
	self.freeLinkid <- linkid
	return
}

func (self *LinkSet) Set(linkid uint16) (ch chan []byte, err error) {
	if !self.isValidLinkid(linkid) {
		err = IndexError
		return
	}

	self.rw.Lock()
	defer self.rw.Unlock()

	if self.chs[linkid] != nil {
		err = ConflictError
		return
	}

	ch = make(chan []byte, 256)
	self.chs[linkid] = ch
	return
}

func (self *LinkSet) Reset(linkid uint16) (err error) {
	if !self.isValidLinkid(linkid) {
		err = IndexError
		return
	}

	self.rw.Lock()
	defer self.rw.Unlock()

	ch := self.chs[linkid]
	if ch != nil {
		close(ch)
		self.chs[linkid] = nil
	} else {
		err = IndexError
	}
	return
}

func (self *LinkSet) PutData(linkid uint16, data []byte) (err error) {
	if !self.isValidLinkid(linkid) {
		err = IndexError
		return
	}

	self.rw.RLock()
	defer self.rw.RUnlock()

	ch := self.chs[linkid]
	if ch != nil {
		ch <- data
	} else {
		err = IndexError
	}
	return
}

func newLinkSet() *LinkSet {
	capacity := options.Capacity
	freeLinkid := make(chan uint16, capacity)
	var i uint16 = 1
	for ; i < capacity; i++ {
		freeLinkid <- i
	}

	linkset := new(LinkSet)
	linkset.capacity = capacity
	linkset.freeLinkid = freeLinkid
	linkset.chs = make([]chan []byte, capacity)
	return linkset
}
