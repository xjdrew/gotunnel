//
//   date  : 2014-06-05
//   author: xjdrew
//

package tunnel

import (
	"sync"
)

var frontServer *FrontServer
var once sync.Once

type FrontDoor struct {
	coor *Coor
}

func (self *FrontDoor) Reload() error {
	return nil
}

func (self *FrontDoor) Start() error {
	frontServer.addCoor(self.coor)
	return self.coor.Start()
}

func (self *FrontDoor) Stop() {
}

func (self *FrontDoor) Wait() {
	self.coor.Wait()
	frontServer.removeCoor(self.coor)
}

func createFrontServer() {
	frontServer = NewFrontServer()
	go func() {
		err := frontServer.Start()
		if err != nil {
			Panic("start front server failed:%s", err.Error())
		}
		frontServer.Wait()
	}()
}

func NewFrontDoor(tunnel *Tunnel) Service {
	once.Do(createFrontServer)

	frontDoor := new(FrontDoor)
	frontDoor.coor = NewCoor(tunnel, nil)
	return frontDoor
}
