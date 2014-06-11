//
//   date  : 2014-06-04
//   author: xjdrew
//

package tunnel

import (
	"errors"
	"sync"
)

var IndexError = errors.New("index out of range")
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

func (self *LinkSet) Set(linkid uint16, ch chan []byte) (err error) {
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
	self.chs[linkid] = ch
	return
}

func (self *LinkSet) Reset(linkid uint16) (ch chan []byte, err error) {
	if !self.isValidLinkid(linkid) {
		err = IndexError
		return
	}
	self.rw.Lock()
	ch = self.chs[linkid]
	self.chs[linkid] = nil
	self.rw.Unlock()
	return
}

func (self *LinkSet) Get(linkid uint16) (ch chan []byte, err error) {
	if !self.isValidLinkid(linkid) {
		err = IndexError
		return
	}

	self.rw.RLock()
	ch = self.chs[linkid]
	self.rw.RUnlock()
	return
}

func NewLinkSet(capacity uint16) LinkSet {
	freeLinkid := make(chan uint16, capacity)
	var i uint16 = 1
	for ; i < capacity; i++ {
		freeLinkid <- i
	}

	var linkset LinkSet
	linkset.capacity = capacity
	linkset.freeLinkid = freeLinkid
	linkset.chs = make([]chan []byte, capacity)
	return linkset
}
