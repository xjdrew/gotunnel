//
//   date  : 2014-07-16
//   author: xjdrew
//

package tunnel

import (
	"net"
	"sync"
)

type TunnelClient struct {
	tunnelAddr string
	wg         sync.WaitGroup
	newDoor    func(*Tunnel) Service
}

func (self *TunnelClient) Start() error {
	addr, err := net.ResolveTCPAddr("tcp", self.tunnelAddr)
	if err != nil {
		return err
	}

	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		return err
	}

	err = writeTGW(conn)
	if err != nil {
		Error("write tgw failed")
		return err
	}

	Info("create tunnel: %v <-> %v", conn.LocalAddr(), conn.RemoteAddr())
	tunnel := NewTunnel(conn)
	door := self.newDoor(tunnel)
	err = door.Start()
	if err != nil {
		Error("door start failed:%s", err.Error())
		return nil
	}
	door.Wait()
	return nil
}

func (self *TunnelClient) Reload() error {
	return nil
}

func (self *TunnelClient) Stop() {
}

func (self *TunnelClient) Wait() {
	self.wg.Wait()
}

func NewTunnelClient(newDoor func(*Tunnel) Service) *TunnelClient {
	tunnelClient := new(TunnelClient)
	tunnelClient.tunnelAddr = options.TunnelAddr
	tunnelClient.newDoor = newDoor
	return tunnelClient
}
