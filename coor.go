//
//   date  : 2014-06-05
//   author: xjdrew
//

package main

import (
    "bytes"
    "encoding/binary"
    "sync"
)

const (
    LINK_DATA     uint8  = iota
    LINK_CREATE  
    LINK_DESTROY 
)

type CmdBody struct {
    Cmd    uint8
    Linkid uint16
}

type Coor struct {
    gate bool
    tunnel *Tunnel
    linkset *LinkSet
    wg sync.WaitGroup
}

func (self *Coor) SetTunnel(tunnel *Tunnel) {
    self.tunnel = tunnel
}

func (self *Coor) Start() error {
    self.wg.Add(1)
    go self.tunnel.PumpOut(&self.wg)
    self.wg.Add(1)
    go self.tunnel.PumpUp(&self.wg)
    self.wg.Add(1)
    go self.Dispatch() 
    return nil
}

func (self *Coor) Wait() {
    self.wg.Wait()
}

func (self *Coor) AddLink(ch chan[]byte) uint16 {
    if self.tunnel == nil {
        return 0
    }
    return self.linkset.Set(ch)
}

func (self *Coor) RemLink(linkid uint16) {
    self.linkset.Clear(linkid)
}

func (self *Coor) Send(cmd uint8, linkid uint16, data []byte) {
    var payload TunnelPayload
    switch cmd {
        case LINK_DATA:
            payload.Linkid = linkid
            payload.Data = data
        case LINK_CREATE, LINK_DESTROY:
            buf := new(bytes.Buffer)
            var body CmdBody
            body.Cmd = cmd
            body.Linkid = linkid
            binary.Write(buf, binary.LittleEndian, &body)

            payload.Linkid = 0
            payload.Data   = buf.Bytes()
    }
    self.tunnel.Put(&payload)
}

func (self *Coor) OnCmd(body *CmdBody) error {
    switch body.Cmd {
        case LINK_CREATE:
        case LINK_DESTROY:
            ch := self.linkset.Get(body.Linkid)
            if ch != nil {
                close(ch)
            }
        default:
            logger.Printf("receive unknown cmd:%v", body)
    }
    return nil
}

func (self *Coor) Dispatch() {
    defer self.wg.Done()
    for {
        payload := self.tunnel.Pop()
        if payload == nil {
            logger.Printf("pop message failed, break dispatch")
            break
        }

        if payload.Linkid == 0 {
            var body CmdBody
            buf := bytes.NewBuffer(payload.Data)
            err := binary.Read(buf, binary.LittleEndian, &body)
            if err != nil {
                logger.Printf("parse message failed:%s, break dispatch", err.Error())
                break
            }
            
            err = self.OnCmd(&body)
            if err != nil {
                logger.Printf("deal cmd failed:%s, break dispatch", err.Error())
                break
            }
        } else {
            ch := self.linkset.Get(payload.Linkid)
            if ch == nil {
                logger.Printf("unknown linkid:%d, drop message", payload.Linkid)
            } else {
                ch <- payload.Data
            }
        }
    }
}

func NewCoor(gate bool, capacity uint16) *Coor {
    linkset := NewLinkSet(capacity)
    var wg sync.WaitGroup
    return &Coor{gate, nil, linkset, wg}
}

