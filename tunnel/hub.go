//
//   date  : 2014-06-05
//   author: xjdrew
//

package tunnel

import (
	"bytes"
	"encoding/binary"
)

const (
	LINK_DATA uint8 = iota
	LINK_CREATE
	LINK_CLOSE
	LINK_CLOSE_RECV
	LINK_CLOSE_SEND
)

type Cmd struct {
	Cmd    uint8
	Linkid uint16
}

type CtrlDelegate interface {
	Ctrl(cmd *Cmd) bool
}

type Hub struct {
	*LinkSet
	tunnel *Tunnel

	delegate CtrlDelegate
}

func (self *Hub) SetCtrlDelegate(delegate CtrlDelegate) {
	self.delegate = delegate
}

func (self *Hub) Send(cmd uint8, linkid uint16, data []byte) bool {
	switch cmd {
	case LINK_DATA:
		Info("link(%d) send %d bytes data", linkid, len(data))
	default:
		buf := bytes.NewBuffer(mpool.Get()[0:0])
		var body Cmd
		body.Cmd = cmd
		body.Linkid = linkid
		binary.Write(buf, binary.LittleEndian, &body)

		linkid = 0
		data = buf.Bytes()
		Info("link(%d) send cmd:%d", linkid, cmd)
	}

	if err := self.tunnel.Write(linkid, data); err != nil {
		Error("link(%d) write to %s failed:%s", linkid, self.tunnel, err.Error())
		return false
	}
	return true
}

func (self *Hub) onCtrl(cmd *Cmd) {
	if self.delegate != nil && self.delegate.Ctrl(cmd) {
		return
	}

	linkid := cmd.Linkid
	link := self.getLink(linkid)
	if link == nil {
		Error("link(%d) recv cmd:%d, no link", linkid, cmd.Cmd)
		return
	}

	switch cmd.Cmd {
	case LINK_CLOSE:
		link.resetRSflag()
	case LINK_CLOSE_RECV:
		link.resetSflag()
	case LINK_CLOSE_SEND:
		link.resetRflag()
	default:
		Error("link(%d) receive unknown cmd:%v", linkid, cmd)
	}
}

func (self *Hub) onData(linkid uint16, data []byte) {
	link := self.getLink(linkid)

	if link == nil {
		mpool.Put(data)
		Error("link(%d) no link", linkid)
		return
	}

	if !link.putData(data) {
		mpool.Put(data)
		Error("link(%d) put data failed", linkid)
		return
	}
	return
}

func (self *Hub) dispatch() {
	defer self.tunnel.Close()

	var cmd Cmd
	for {
		linkid, data, err := self.tunnel.Read()
		if err != nil {
			Error("%s read failed:%v", self.tunnel, err)
			break
		}

		if linkid == 0 {
			buf := bytes.NewBuffer(data)
			err := binary.Read(buf, binary.LittleEndian, &cmd)
			mpool.Put(data)
			if err != nil {
				Error("parse message failed:%s, break dispatch", err.Error())
				break
			}
			Info("link(%d) recv cmd:%d", cmd.Linkid, cmd.Cmd)
			self.onCtrl(&cmd)
		} else {
			Info("link(%d) recv %d bytes data", linkid, len(data))
			self.onData(linkid, data)
		}
	}
}

func (self *Hub) Start() {
	self.dispatch()

	// tunnel disconnect, so reset all link
	Error("reset all link")
	for i := uint16(1); i < MaxLinkPerTunnel; i++ {
		link := self.getLink(i)
		if link != nil {
			link.resetRSflag()
			Error("link(%d) reset", i)
		}
	}
	Log("hub(%s) quit", self.tunnel)
}

func (self *Hub) Status() {
	total := 0
	links := make([]uint16, 100)
	for i := uint16(0); i < MaxLinkPerTunnel; i++ {
		if self.links[i] != nil {
			if total < cap(links) {
				links[total] = i
			}
			total += 1
		}
	}
	if total <= cap(links) {
		links = links[:total]
	}
	Log("<status> %s, %d links(%v)", self.tunnel, total, links)
}

func (self *Hub) NewLink(linkid uint16) *Link {
	link := newLink(linkid, self)
	if self.setLink(linkid, link) {
		return link
	}
	return nil
}

func (self *Hub) ReleaseLink(linkid uint16) bool {
	return self.resetLink(linkid)
}

func newHub(tunnel *Tunnel, client bool) *Hub {
	hub := new(Hub)
	hub.LinkSet = newLinkSet(client)
	hub.tunnel = tunnel
	return hub
}
