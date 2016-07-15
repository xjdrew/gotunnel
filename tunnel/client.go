//
//   date  : 2014-07-16
//   author: xjdrew
//

package tunnel

import (
	"container/heap"
	"errors"
	"net"
	"sync"
	"time"
)

// client hub
type ClientHub struct {
	*Hub
	sent uint16
	rcvd uint16
}

func (h *ClientHub) heartbeat() {
	c := time.Tick(1 * time.Second)

	timeout := Timeout
	if Timeout <= 0 {
		timeout = TunnelMaxTimeout
	}
	for range c {
		// id overflow
		span := h.sent - h.rcvd
		if int(span) >= timeout {
			Error("tunnel(%v) timeout, sent:%d, rcvd:%d", h.Hub.tunnel, h.sent, h.rcvd)
			h.Hub.Close()
			break
		}

		h.sent = h.sent + 1
		if !h.SendCmd(h.sent, TUN_HEARTBEAT) {
			break
		}
	}
}

func (h *ClientHub) onCtrl(cmd Cmd) bool {
	id := cmd.Id
	switch cmd.Cmd {
	case TUN_HEARTBEAT:
		h.rcvd = id
		return true
	}
	return false
}

func newClientHub(tunnel *Tunnel) *ClientHub {
	h := &ClientHub{
		Hub: newHub(tunnel),
	}
	h.Hub.onCtrlFilter = h.onCtrl
	go h.heartbeat()
	return h
}

// tunnel client
type Client struct {
	laddr   string
	backend string
	secret  string
	tunnels uint

	alloc *IdAllocator
	cq    HubQueue
	lock  sync.Mutex
}

func (cli *Client) createHub() (hub *HubItem, err error) {
	conn, err := dial(cli.backend)
	if err != nil {
		return
	}

	tunnel := newTunnel(conn)
	_, challenge, err := tunnel.ReadPacket()
	if err != nil {
		Error("read challenge failed(%v):%s", tunnel, err)
		return
	}

	a := NewTaa(cli.secret)
	token, ok := a.ExchangeCipherBlock(challenge)
	if !ok {
		err = errors.New("exchange chanllenge failed")
		Error("exchange challenge failed(%v)", tunnel)
		return
	}

	if err = tunnel.WritePacket(0, token); err != nil {
		Error("write token failed(%v):%s", tunnel, err)
		return
	}

	tunnel.SetCipherKey(a.GetRc4key())
	hub = &HubItem{
		ClientHub: newClientHub(tunnel),
	}
	return
}

func (cli *Client) addHub(item *HubItem) {
	cli.lock.Lock()
	heap.Push(&cli.cq, item)
	cli.lock.Unlock()
}

func (cli *Client) removeHub(item *HubItem) {
	cli.lock.Lock()
	heap.Remove(&cli.cq, item.index)
	cli.lock.Unlock()
}

func (cli *Client) fetchHub() *HubItem {
	defer cli.lock.Unlock()
	cli.lock.Lock()

	if len(cli.cq) == 0 {
		return nil
	}
	item := cli.cq[0]
	item.priority += 1
	heap.Fix(&cli.cq, 0)
	return item
}

func (cli *Client) dropHub(item *HubItem) {
	cli.lock.Lock()
	item.priority -= 1
	heap.Fix(&cli.cq, item.index)
	cli.lock.Unlock()
}

func (cli *Client) handleConn(hub *HubItem, conn *net.TCPConn) {
	defer Recover()
	defer cli.dropHub(hub)
	defer conn.Close()

	id := cli.alloc.Acquire()
	defer cli.alloc.Release(id)

	h := hub.Hub
	l := h.createLink(id)
	defer h.deleteLink(id)

	h.SendCmd(id, LINK_CREATE)
	h.startLink(l, conn)
}

func (cli *Client) listen() error {
	ln, err := net.Listen("tcp", cli.laddr)
	if err != nil {
		return err
	}

	defer ln.Close()

	tcpListener := ln.(*net.TCPListener)
	for {
		conn, err := tcpListener.AcceptTCP()
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Temporary() {
				Log("acceept failed temporary: %s", netErr.Error())
				continue
			} else {
				return err
			}
		}
		Info("new connection from %v", conn.RemoteAddr())
		hub := cli.fetchHub()
		if hub == nil {
			Error("no active hub")
			conn.Close()
			continue
		}

		conn.SetKeepAlive(true)
		conn.SetKeepAlivePeriod(time.Second * 60)
		go cli.handleConn(hub, conn)
	}
}

func (cli *Client) Start() error {
	sz := cap(cli.cq)
	for i := 0; i < sz; i++ {
		go func(index int) {
			Recover()

			for {
				hub, err := cli.createHub()
				if err != nil {
					Error("tunnel %d reconnect failed", index)
					time.Sleep(time.Second * 3)
					continue
				}

				Error("tunnel %d connect succeed", index)
				cli.addHub(hub)
				hub.Start()
				cli.removeHub(hub)
				Error("tunnel %d disconnected", index)
			}
		}(i)
	}

	return cli.listen()
}

func (cli *Client) Status() {
	defer cli.lock.Unlock()
	cli.lock.Lock()
	for _, hub := range cli.cq {
		hub.Status()
	}
}

func NewClient(listen, backend, secret string, tunnels uint) (*Client, error) {
	client := &Client{
		laddr:   listen,
		backend: backend,
		secret:  secret,
		tunnels: tunnels,

		alloc: newAllocator(),
		cq:    make(HubQueue, tunnels)[0:0],
	}
	return client, nil
}
