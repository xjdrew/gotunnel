//
//   date  : 2014-07-16
//   author: xjdrew
//

package tunnel

import (
	"errors"
	"io"
	"net"
	"sync"
	"time"
)

type Client struct {
	app  *App
	hubs []*Hub
	off  int // current hub
	wg   sync.WaitGroup
}

func (cli *Client) createHub() (hub *Hub, err error) {
	conn, err := net.DialTCP("tcp", nil, cli.app.baddr)
	if err != nil {
		return
	}
	Info("create tunnel: %v <-> %v", conn.LocalAddr(), conn.RemoteAddr())

	// auth
	challenge := make([]byte, TaaBlockSize)
	if _, err = io.ReadFull(conn, challenge); err != nil {
		Error("read challenge failed(%v):%s", conn.RemoteAddr(), err)
		return
	}
	Debug("challenge(%v), len %d, %v", conn.RemoteAddr(), len(challenge), challenge)

	a := NewTaa(cli.app.Secret)
	token, ok := a.ExchangeCipherBlock(challenge)
	if !ok {
		err = errors.New("exchange chanllenge failed")
		Error("exchange challenge failed(%v)", conn.RemoteAddr())
		return
	}

	Debug("token(%v), len %d, %v", conn.RemoteAddr(), len(token), token)
	if _, err = conn.Write(token); err != nil {
		Error("write token failed(%v):%s", conn.RemoteAddr(), err)
		return
	}

	hub = newHub(newTunnel(conn, a.GetRc4key()), true)
	return
}

func (cli *Client) handleConn(hub *Hub, conn BiConn) {
	defer conn.Close()
	defer Recover()

	linkid := hub.AcquireId()
	if linkid == 0 {
		Error("alloc linkid failed, source: %v", conn.RemoteAddr())
		return
	}
	defer hub.ReleaseId(linkid)

	Info("link(%d) create link, source: %v", linkid, conn.RemoteAddr())
	link := hub.NewLink(linkid)
	defer hub.ReleaseLink(linkid)

	link.SendCreate()
	link.Pump(conn)
}

func (cli *Client) fetchHub() *Hub {
	for i := 0; i < len(cli.hubs); i++ {
		hub := cli.hubs[cli.off]
		cli.off = (cli.off + 1) % len(cli.hubs)
		if hub != nil {
			return hub
		}
	}
	return nil
}

func (cli *Client) listen() {
	defer cli.wg.Done()

	ln, err := net.ListenTCP("tcp", cli.app.laddr)
	if err != nil {
		Panic("listen failed:%v", err)
	}

	for {
		conn, err := ln.AcceptTCP()
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
	done := make(chan error, len(cli.hubs))
	for i := 0; i < len(cli.hubs); i++ {
		go func(index int) {
			Recover()

			first := true
			for {
				hub, err := cli.createHub()
				if first {
					first = false
					done <- err
					if err != nil {
						Error("tunnel %d connect failed", index)
						break
					}
				} else if err != nil {
					Error("tunnel %d reconnect failed", index)
					time.Sleep(time.Second)
					continue
				}

				Error("tunnel %d connect succeed", index)
				cli.hubs[index] = hub
				hub.Start()
				cli.hubs[index] = nil
				Error("tunnel %d disconnected", index)
			}
		}(i)
	}

	for i := 0; i < len(cli.hubs); i++ {
		err := <-done
		if err != nil {
			return err
		}
	}

	cli.wg.Add(1)
	go cli.listen()
	return nil
}

func (cli *Client) Wait() {
	cli.wg.Wait()
	Log("tunnel client quit")
}

func (cli *Client) Status() {
	for _, hub := range cli.hubs {
		hub.Status()
	}
}

func newClient(app *App) *Client {
	return &Client{
		app:  app,
		hubs: make([]*Hub, app.Tunnels),
	}
}
