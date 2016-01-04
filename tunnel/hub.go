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
	TUN_HEARTBEAT
)

type Cmd struct {
	Cmd uint8  // control command
	Id  uint16 // id
}

type Hub struct {
	tunnel *Tunnel

	ll    sync.RWMutex // protect links
	links map[uint16]*link

	onCtrlFilter func(cmd Cmd) bool
}

func (h *Hub) SendCmd(id uint16, cmd uint8) bool {
	buf := bytes.NewBuffer(mpool.Get()[0:0])
	c := Cmd{
		Cmd: cmd,
		Id:  id,
	}
	binary.Write(buf, binary.LittleEndian, &c)
	Info("link(%d) send cmd:%d", id, cmd)
	return h.Send(0, buf.Bytes())
}

func (h *Hub) Send(id uint16, data []byte) bool {
	if err := h.tunnel.WritePacket(id, data); err != nil {
		Error("link(%d) write to %s failed:%s", id, h.tunnel, err.Error())
		return false
	}
	return true
}

func (h *Hub) onCtrl(cmd Cmd) {
	if h.onCtrlFilter != nil && h.onCtrlFilter(cmd) {
		return
	}

	id := cmd.Id
	l := h.getLink(id)
	if l == nil {
		Error("link(%d) recv cmd:%d, no link", id, cmd.Cmd)
		return
	}

	switch cmd.Cmd {
	case LINK_CLOSE:
		l.aclose()
	case LINK_CLOSE_RECV:
		l.rclose()
	case LINK_CLOSE_SEND:
		l.wclose()
	default:
		Error("link(%d) receive unknown cmd:%v", id, cmd)
	}
}

func (h *Hub) onData(id uint16, data []byte) {
	link := h.getLink(id)

	if link == nil {
		mpool.Put(data)
		Error("link(%d) no link", id)
		return
	}

	if !link.write(data) {
		mpool.Put(data)
		Error("link(%d) put data failed", id)
		return
	}
	return
}

func (h *Hub) Start() {
	defer h.tunnel.Close()

	for {
		id, data, err := h.tunnel.ReadPacket()
		if err != nil {
			Error("%s read failed:%v", h.tunnel, err)
			break
		}

		if id == 0 {
			var cmd Cmd
			buf := bytes.NewBuffer(data)
			err := binary.Read(buf, binary.LittleEndian, &cmd)
			mpool.Put(data)
			if err != nil {
				Error("parse message failed:%s, break dispatch", err.Error())
				break
			}
			Info("link(%d) recv cmd:%d", cmd.Id, cmd.Cmd)
			h.onCtrl(cmd)
		} else {
			Info("link(%d) recv %d bytes data", id, len(data))
			h.onData(id, data)
		}
	}

	// tunnel disconnect, so reset all link
	Error("reset all link")
	h.resetAllLink()
	Log("hub(%s) quit", h.tunnel)
}

func (h *Hub) Close() {
	h.tunnel.Close()
}

func (h *Hub) Status() {
	h.ll.RLock()
	defer h.ll.RUnlock()
	var links []uint16
	for id := range h.links {
		links = append(links, id)
	}
	Log("<status> %s, %d links(%v)", h.tunnel, len(h.links), links)
}

func (h *Hub) resetAllLink() {
	h.ll.RLock()
	defer h.ll.RUnlock()

	for _, l := range h.links {
		l.aclose()
	}
}

func newHub(tunnel *Tunnel) *Hub {
	return &Hub{
		tunnel: tunnel,
		links:  make(map[uint16]*link),
	}
}
