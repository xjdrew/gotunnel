//
//   date  : 2014-06-04
//   author: xjdrew
//
package tunnel

type LinkSet struct {
	capacity   uint16
	freeLinkid chan uint16
	links      []*Link
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

func (self *LinkSet) setLink(id uint16, link *Link) bool {
	if self.links[id] != nil {
		return false
	}
	self.links[id] = link
	return true
}

func (self *LinkSet) getLink(id uint16) *Link {
	return self.links[id]
}

func (self *LinkSet) resetLink(id uint16) bool {
	if self.links[id] != nil {
		self.links[id] = nil
		return true
	}
	return false
}

func newLinkSet(capacity uint16) *LinkSet {
	linkset := new(LinkSet)
	linkset.capacity = capacity
	linkset.links = make([]*Link, capacity)
	if options.Server != "" {
		freeLinkid := make(chan uint16, capacity)
		for i := uint16(1); i < capacity; i++ {
			freeLinkid <- i
		}
		linkset.freeLinkid = freeLinkid
	}

	return linkset
}
