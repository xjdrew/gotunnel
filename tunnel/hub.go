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
	LINK_DESTROY
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
	tunnel   *Tunnel
	delegate CtrlDelegate
	wg       sync.WaitGroup
}

func (self *Hub) SetCtrlDelegate(delegate CtrlDelegate) {
	self.delegate = delegate
}

func (self *Hub) SendLinkCreate(linkid uint16) {
	self.Send(LINK_CREATE, linkid, nil)
}

func (self *Hub) SendLinkDestory(linkid uint16) {
	self.Send(LINK_DESTROY, linkid, nil)
}

func (self *Hub) SendLinkData(linkid uint16, data []byte) {
	self.Send(LINK_DATA, linkid, data)
}

func (self *Hub) Send(cmd uint8, linkid uint16, data []byte) {
	payload := new(TunnelPayload)
	switch cmd {
	case LINK_DATA:
		Debug("link(%d) send data:%d", linkid, len(data))

		payload.Linkid = linkid
		payload.Data = data
	case LINK_CREATE, LINK_DESTROY:
		Debug("link(%d) send cmd:%d", linkid, cmd)

		buf := new(bytes.Buffer)
		var body CmdPayload
		body.Cmd = cmd
		body.Linkid = linkid
		binary.Write(buf, binary.LittleEndian, &body)

		payload.Linkid = 0
		payload.Data = buf.Bytes()
	default:
		Error("unknown cmd:%d, linkid:%d", cmd, linkid)
	}
	self.tunnel.Put(payload)
}

func (self *Hub) ctrl(cmd *CmdPayload) {
	linkid := cmd.Linkid
	Debug("link(%d) recv cmd:%d", linkid, cmd.Cmd)

	if self.delegate != nil && self.delegate.Ctrl(cmd) {
		return
	}

	switch cmd.Cmd {
	case LINK_DESTROY:
		err := self.Reset(linkid)
		if err != nil {
			Info("link(%d) close failed: %v", linkid, err)
		} else {
			Info("link(%d) closed", linkid)
		}
	default:
		Error("receive unknown cmd:%v", cmd)
	}
}

func (self *Hub) data(payload *TunnelPayload) {
	linkid := payload.Linkid
	Debug("link(%d) recv data:%d", linkid, len(payload.Data))

	err := self.PutData(linkid, payload.Data)
	if err != nil {
		Error("link(%d) put data failed: %v", linkid, err)
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

func (self *Hub) pumpUp() {
	self.wg.Done()
	defer Recover()

	self.tunnel.PumpUp()
}

func (self *Hub) Start() error {
	self.wg.Add(1)
	go self.pumpOut()
	self.wg.Add(1)
	go self.pumpUp()
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
		err := self.Reset(i)
		if err == nil {
			Info("link(%d) closed", i)
		}
	}
	Log("hub(%s) quit", self.tunnel.String())
}

func newHub(tunnel *Tunnel) *Hub {
	hub := new(Hub)
	hub.LinkSet = newLinkSet()
	hub.tunnel = tunnel
	return hub
}
