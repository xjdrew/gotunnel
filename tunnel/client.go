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

type Client struct {
	laddr   string
	backend string
	secret  string
	tunnels uint

	alloc *IdAllocator
	cq    HubQueue
	lock  sync.Mutex
}

const (
	dailTimeoutSeconds = 5 * time.Second
)

func (cli *Client) createHub() (hub *HubItem, err error) {
	conn, err := net.DialTimeout("tcp", cli.backend, dailTimeoutSeconds)
	if err != nil {
		return
	}

	tunnel := newTunnel(conn.(*net.TCPConn))
	_, challenge, err := tunnel.Read()
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

	if err = tunnel.Write(0, token); err != nil {
		Error("write token failed(%v):%s", tunnel, err)
		return
	}

	tunnel.SetCipherKey(a.GetRc4key())
	hub = &HubItem{
		Hub: newHub(tunnel),
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
	defer conn.Close()
	defer Recover()
	defer cli.dropHub(hub)

	linkid := cli.alloc.Acquire()
	defer cli.alloc.Release(linkid)

	Info("link(%d) create link, source: %v", linkid, conn.RemoteAddr())
	link := newLink(linkid, hub.Hub)

	link.SendCreate()
	link.Pump(conn)
}

func (cli *Client) listen() {
	ln, err := net.Listen("tcp", cli.laddr)
	if err != nil {
		Panic("listen failed:%v", err)
	}

	tcpListener := ln.(*net.TCPListener)
	for {
		conn, err := tcpListener.AcceptTCP()
		if err != nil {
			Log("acceept failed:%s", err.Error())
			if opErr, ok := err.(*net.OpError); ok {
				if !opErr.Temporary() {
					break
				}
			}
			continue
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

	go cli.listen()
	return nil
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
