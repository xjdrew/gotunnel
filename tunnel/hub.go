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
)

type Cmd struct {
	Cmd    uint8
	Linkid uint16
}

type Hub struct {
	tunnel   *Tunnel
	links    map[uint16]*Link
	linkLock sync.RWMutex

	onCtrlFilter func(cmd Cmd) bool
}

func (hub *Hub) Send(cmd uint8, linkid uint16, data []byte) bool {
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

	if err := hub.tunnel.Write(linkid, data); err != nil {
		Error("link(%d) write to %s failed:%s", linkid, hub.tunnel, err.Error())
		return false
	}
	return true
}

func (hub *Hub) onCtrl(cmd Cmd) {
	if hub.onCtrlFilter != nil && hub.onCtrlFilter(cmd) {
		return
	}

	linkid := cmd.Linkid
	link := hub.getLink(linkid)
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

func (hub *Hub) onData(linkid uint16, data []byte) {
	link := hub.getLink(linkid)

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

func (hub *Hub) Start() {
	defer hub.tunnel.Close()

	for {
		linkid, data, err := hub.tunnel.Read()
		if err != nil {
			Error("%s read failed:%v", hub.tunnel, err)
			break
		}

		if linkid == 0 {
			var cmd Cmd
			buf := bytes.NewBuffer(data)
			err := binary.Read(buf, binary.LittleEndian, &cmd)
			mpool.Put(data)
			if err != nil {
				Error("parse message failed:%s, break dispatch", err.Error())
				break
			}
			Info("link(%d) recv cmd:%d", cmd.Linkid, cmd.Cmd)
			hub.onCtrl(cmd)
		} else {
			Info("link(%d) recv %d bytes data", linkid, len(data))
			hub.onData(linkid, data)
		}
	}

	// tunnel disconnect, so reset all link
	Error("reset all link")
	hub.resetAllLink()
	Log("hub(%s) quit", hub.tunnel)
}

func (hub *Hub) Status() {
	hub.linkLock.RLock()
	defer hub.linkLock.RUnlock()
	var links []uint16
	for linkid := range hub.links {
		links = append(links, linkid)
	}
	Log("<status> %s, %d links(%v)", hub.tunnel, len(hub.links), links)
}

func (hub *Hub) resetAllLink() {
	hub.linkLock.RLock()
	defer hub.linkLock.RUnlock()

	for _, link := range hub.links {
		link.resetRSflag()
	}
}

func (hub *Hub) getLink(linkid uint16) *Link {
	hub.linkLock.RLock()
	defer hub.linkLock.RUnlock()
	return hub.links[linkid]
}

func (hub *Hub) deleteLink(linkid uint16) {
	hub.linkLock.Lock()
	defer hub.linkLock.Unlock()
	delete(hub.links, linkid)
}

func (hub *Hub) setLink(linkid uint16, link *Link) bool {
	hub.linkLock.Lock()
	defer hub.linkLock.Unlock()
	if _, ok := hub.links[linkid]; ok {
		return false
	}
	hub.links[linkid] = link
	return true
}

func newHub(tunnel *Tunnel) *Hub {
	return &Hub{
		tunnel: tunnel,
		links:  make(map[uint16]*Link),
	}
}
