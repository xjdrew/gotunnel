//
//   date  : 2014-06-04
//   author: xjdrew
//

package tunnel

type LinkSet struct {
	capacity   uint16
	rchs       []chan []byte //read channel
	wflags     []bool        //write flag
	freeLinkid chan uint16
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

func (self *LinkSet) ReleaseId(linkid uint16) {
	self.freeLinkid <- linkid
}

func (self *LinkSet) setRWflag(linkid uint16) bool {
	if self.rchs[linkid] != nil || self.wflags[linkid] {
		return false
	}
	self.rchs[linkid] = make(chan []byte, 256)
	self.wflags[linkid] = true
	return true
}

// stop write data to remote
func (self *LinkSet) resetWflag(linkid uint16) bool {
	flag := self.wflags[linkid]
	self.wflags[linkid] = false
	return flag
}

func (self *LinkSet) getWflag(linkid uint16) bool {
	return self.wflags[linkid]
}

// stop recv data from remote
func (self *LinkSet) resetRflag(linkid uint16, dropall bool) bool {
	ch := self.rchs[linkid]
	if ch == nil {
		return false
	}
	self.rchs[linkid] = nil

	if dropall {
		// drop all pending data
	Loop:
		for {
			select {
			case <-ch:
			default:
				break Loop
			}
		}
	}
	close(ch)
	return true
}

func (self *LinkSet) resetRWflag(linkid uint16) bool {
	ok1 := self.resetWflag(linkid)
	ok2 := self.resetRflag(linkid, true)
	return ok1 || ok2
}

func (self *LinkSet) putData(linkid uint16, data []byte) bool {
	ch := self.rchs[linkid]
	if ch == nil {
		return false
	}

	ch <- data
	return true
}

func (self *LinkSet) getDataReader(linkid uint16) func() ([]byte, bool) {
	ch := self.rchs[linkid]
	return func() ([]byte, bool) {
		data, ok := <-ch
		return data, ok
	}
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
	linkset.rchs = make([]chan []byte, capacity)
	linkset.wflags = make([]bool, capacity)
	return linkset
}
