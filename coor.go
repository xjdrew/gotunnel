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

type CmdPayload struct {
    Cmd    uint8
    Linkid uint16
}

type Coor struct {
    tunnel *Tunnel
    outCh chan interface {}
    wg sync.WaitGroup
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

func (self *Coor) SendLinkCreate(linkid uint16) {
    self.Send(LINK_CREATE, linkid, nil)
}

func (self *Coor) SendLinkDestory(linkid uint16) {
    self.Send(LINK_DESTROY, linkid, nil)
}

func (self *Coor) SendLinkData(linkid uint16, data []byte) {
    self.Send(LINK_DATA, linkid, data)
}

func (self *Coor) Send(cmd uint8, linkid uint16, data []byte) {
    var payload TunnelPayload
    switch cmd {
        case LINK_DATA:
            payload.Linkid = linkid
            payload.Data = data
        case LINK_CREATE, LINK_DESTROY:
            buf := new(bytes.Buffer)
            var body CmdPayload
            body.Cmd = cmd
            body.Linkid = linkid
            binary.Write(buf, binary.LittleEndian, &body)

            payload.Linkid = 0
            payload.Data   = buf.Bytes()
        default:
            logger.Printf("unknown cmd:%d, linkid:%d", cmd, linkid)
    }
    self.tunnel.Put(&payload)
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
            var cmd CmdPayload
            buf := bytes.NewBuffer(payload.Data)
            err := binary.Read(buf, binary.LittleEndian, &cmd)
            if err != nil {
                logger.Printf("parse message failed:%s, break dispatch", err.Error())
                break
            }
            self.outCh <- cmd
        } else {
            self.outCh <- payload
        }
    }
}

func NewCoor(tunnel *Tunnel, ch chan interface{}) *Coor {
    coor := new(Coor)
    coor.tunnel = tunnel
    coor.outCh = ch
    return coor
}

