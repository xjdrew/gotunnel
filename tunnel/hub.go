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
	LINK_RECVBUF_HW // recv buffer enter high water
	LINK_RECVBUF_LW // recv buffer enter low water
)

type CmdPayload struct {
	Cmd    uint8
	Linkid uint16
}

type CtrlDelegate interface {
	Ctrl(cmd *CmdPayload) bool
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
	payload := new(TunnelPayload)
	switch cmd {
	case LINK_DATA:
		Info("link(%d) send %d bytes data", linkid, len(data))

		payload.Linkid = linkid
		payload.Data = data
	default:
		Info("link(%d) send cmd:%d", linkid, cmd)

		buf := bytes.NewBuffer(mpool.Get()[0:0])
		var body CmdPayload
		body.Cmd = cmd
		body.Linkid = linkid
		binary.Write(buf, binary.LittleEndian, &body)

		payload.Linkid = 0
		payload.Data = buf.Bytes()
	}
	err := self.tunnel.Write(payload)
	if err != nil {
		Error("%s write failed:%v", self.tunnel.String(), err)
		return false
	}
	return true
}

func (self *Hub) ctrl(cmd *CmdPayload) {
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
	case LINK_RECVBUF_HW:
		link.setRemoteQosFlag(true)
	case LINK_RECVBUF_LW:
		link.setRemoteQosFlag(false)
	default:
		Error("link(%d) receive unknown cmd:%v", linkid, cmd)
	}
}

func (self *Hub) data(payload *TunnelPayload) {
	linkid := payload.Linkid
	link := self.getLink(linkid)

	if link == nil {
		mpool.Put(payload.Data)
		Error("link(%d) no link", linkid)
		return
	}

	if !link.putData(payload.Data) {
		mpool.Put(payload.Data)
		Error("link(%d) put data failed", linkid)
		return
	}
	return
}

func (self *Hub) dispatch() {
	defer self.tunnel.Close()

	for {
		payload, err := self.tunnel.Read()
		if err != nil {
			Error("%s read failed:%v", self.tunnel.String(), err)
			break
		}

		if payload.Linkid == 0 {
			cmd := new(CmdPayload)
			buf := bytes.NewBuffer(payload.Data)
			err := binary.Read(buf, binary.LittleEndian, cmd)
			mpool.Put(payload.Data)
			if err != nil {
				Error("parse message failed:%s, break dispatch", err.Error())
				break
			}
			Info("link(%d) recv cmd:%d", cmd.Linkid, cmd.Cmd)
			self.ctrl(cmd)
		} else {
			Info("link(%d) recv %d bytes data", payload.Linkid, len(payload.Data))
			self.data(payload)
		}
	}
}

func (self *Hub) Start() {
	self.dispatch()

	// tunnel disconnect, so reset all link
	Info("reset all link")
	for i := uint16(1); i < self.LinkSet.capacity; i++ {
		link := self.getLink(i)
		if link != nil {
			link.resetRSflag()
		}
	}
	Log("hub(%s) quit", self.tunnel.String())
}

func (self *Hub) Status() {
	total := 0
	links := make([]uint16, 100)
	for i := uint16(0); i < self.capacity; i++ {
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
	Log("<status> %s, %d links(%v)", self.tunnel.String(), total, links)
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

func newHub(tunnel *Tunnel) *Hub {
	hub := new(Hub)
	hub.LinkSet = newLinkSet(uint16(options.Capacity))
	hub.tunnel = tunnel
	return hub
}
