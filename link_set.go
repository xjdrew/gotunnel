//
//   date  : 2014-06-04
//   author: xjdrew
//

package main

type LinkSet struct {
    capacity uint16
    linkChs []chan []byte
    linkidCh chan uint16
}

func (self *LinkSet) Set(linkCh chan []byte) uint16 {
    var linkid uint16
    select {
        case linkid = <- self.linkidCh:
            self.linkChs[linkid] = linkCh
        default:
            logger.Printf("allocate linkid failed")
            linkid = 0
    }
    return linkid
}

func (self *LinkSet) Get(linkid uint16) chan []byte {
    if linkid == 0 || linkid > self.capacity {
        return nil
    }
    return self.linkChs[linkid]
}

func (self *LinkSet) Clear(linkid uint16) {
    if self.Get(linkid) != nil {
        self.linkChs[linkid] = nil
        self.linkidCh <- linkid
    }
}

func NewLinkSet(capacity uint16) *LinkSet {
    linkidCh := make(chan uint16, capacity)
    var i uint16 = 1 
    for ; i < capacity; i++ {
        logger.Printf("alloc id: %d", i)
        linkidCh <- i
    }
    logger.Printf("capacity:%d", capacity)
    return &LinkSet{capacity, make([]chan []byte, capacity), linkidCh}
}

