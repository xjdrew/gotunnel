//
//   date  : 2014-06-04
//   author: xjdrew
//

package main
import (
    "net"
    "encoding/binary"
    "sync"
)

type TunnelPayload struct {
    Linkid uint16
    Data   []byte
}

type Tunnel struct {
    inputCh chan TunnelPayload
    outputCh chan TunnelPayload
    conn *net.TCPConn
}

func (self *Tunnel) Put(payload *TunnelPayload) {
    self.inputCh <- *payload
}

func (self *Tunnel) Pop() *TunnelPayload {
    payload,ok := <- self.outputCh
    if !ok {
        return nil
    }
    return &payload
}

// read
func (self *Tunnel) PumpOut(wg *sync.WaitGroup) {
    defer wg.Done()

    var header struct{Linkid uint16; Sz uint8}
    for {
        err := binary.Read(self.conn, binary.LittleEndian, &header)
        if err != nil {
            logger.Printf("read tunnel failed:%s", err.Error())
            close(self.outputCh)
            break
        }

        var data []byte
        
        // if header.Sz == 0, it's ok too
        data = make([]byte, header.Sz)
        c := 0
        for c < int(header.Sz) {
            n, err := self.conn.Read(data[c:])
            if err != nil {
                logger.Printf("read tunnel failed:%s", err.Error())
                break
            }
            c += n
        }
        
        self.outputCh <- TunnelPayload{header.Linkid, data}
    }
}

// write
func (self *Tunnel) PumpUp(wg *sync.WaitGroup) {
    defer wg.Done()

    var header struct{Linkid uint16; Sz uint8}
    for {
        payload := <- self.inputCh
        
        sz := len(payload.Data)
        if sz > 0xff {
            logger.Panicf("receive malformed payload, linkid:%d, sz:%d", payload.Linkid, sz)
            break
        }

        header.Linkid = payload.Linkid
        header.Sz     = uint8(sz)
        err := binary.Write(self.conn, binary.LittleEndian, &header)
        if err != nil {
            logger.Printf("write tunnel failed:%s", err.Error())
            break
        }
        
        c := 0
        for c < sz {
            n, err := self.conn.Write(payload.Data[c:])
            if err != nil {
                logger.Printf("write tunnel failed:%s", err.Error())
                break
            }
            c += n
        }
    }
}

func NewTunnel(conn *net.TCPConn) *Tunnel {
    return &Tunnel{make(chan TunnelPayload, 65535), make(chan TunnelPayload, 65535), conn}
}

