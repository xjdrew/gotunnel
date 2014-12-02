//
//   date  : 2014-06-05
//   author: xjdrew
//

package tunnel

import (
	"bytes"
	"encoding/binary"
	"sync"
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
	wg       sync.WaitGroup
}

func (self *Hub) SetCtrlDelegate(delegate CtrlDelegate) {
	self.delegate = delegate
}

func (self *Hub) Send(cmd uint8, linkid uint16, data []byte) {
	payload := new(TunnelPayload)
	switch cmd {
	case LINK_DATA:
		Debug("link(%d) send data:%d", linkid, len(data))

		payload.Linkid = linkid
		payload.Data = data
	default:
		Debug("link(%d) send cmd:%d", linkid, cmd)

		buf := new(bytes.Buffer)
		var body CmdPayload
		body.Cmd = cmd
		body.Linkid = linkid
		binary.Write(buf, binary.LittleEndian, &body)

		payload.Linkid = 0
		payload.Data = buf.Bytes()
	}
	self.tunnel.Put(payload)
}

func (self *Hub) ctrl(cmd *CmdPayload) {
	linkid := cmd.Linkid
	Info("link(%d) recv cmd:%d", linkid, cmd.Cmd)

	if self.delegate != nil && self.delegate.Ctrl(cmd) {
		return
	}

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
	Debug("link(%d) recv data:%d", linkid, len(payload.Data))

	link := self.getLink(linkid)
	if link == nil {
		Error("link(%d) no link", linkid)
		return
	}

	if !link.putData(payload.Data) {
		Error("link(%d) put data failed", linkid)
	}
	return
}

func (self *Hub) dispatch() {
	defer self.wg.Done()
	defer Recover()

	for {
		payload := self.tunnel.Pop()
		if payload == nil {
			Error("pop message failed, break dispatch")
			break
		}

		if payload.Linkid == 0 {
			cmd := new(CmdPayload)
			buf := bytes.NewBuffer(payload.Data)
			err := binary.Read(buf, binary.LittleEndian, cmd)
			if err != nil {
				Error("parse message failed:%s, break dispatch", err.Error())
				break
			}
			self.ctrl(cmd)
		} else {
			self.data(payload)
		}
	}
}

func (self *Hub) pumpOut() {
	self.wg.Done()
	defer Recover()

	self.tunnel.PumpOut()
}

func (self *Hub) pumpIn() {
	self.wg.Done()
	defer Recover()

	self.tunnel.PumpIn()
}

func (self *Hub) Start() error {
	self.wg.Add(1)
	go self.pumpOut()

	self.wg.Add(1)
	go self.pumpIn()

	self.wg.Add(1)
	go self.dispatch()
	return nil
}

func (self *Hub) Close() {
	self.tunnel.Close()
}

func (self *Hub) Wait() {
	self.wg.Wait()
	// tunnel disconnect, so reset all link
	Info("reset all link")
	var i uint16 = 1
	for ; i < options.Capacity; i++ {
		link := self.getLink(i)
		if link != nil {
			link.resetSflag()
		}
	}
	Log("hub(%s) quit", self.tunnel.String())
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
	hub.LinkSet = newLinkSet()
	hub.tunnel = tunnel
	return hub
}
