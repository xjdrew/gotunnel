//
//   date  : 2014-12-02
//   author: xjdrew
//

package tunnel

import "sync"

type Qos struct {
	highWater        int
	lowWater         int
	localFlag        bool       // 本端高水位标志
	remoteFlag       int        // 0:normal, 1:对端低水位标志, -1: close
	cond             *sync.Cond // 对端低水位通知
	enterHighWaterCb func()
	enterLowWaterCb  func()
}

func (q *Qos) SetWater(water int) {
	if q.localFlag && water <= q.lowWater {
		q.localFlag = false
		q.enterLowWaterCb()
		return
	}

	if !q.localFlag && water >= q.highWater {
		q.localFlag = true
		q.enterHighWaterCb()
	}
}

func (q *Qos) SetRemoteFlag(flag bool) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	if q.remoteFlag == -1 {
		return
	}
	if flag {
		q.remoteFlag = 1
	} else {
		q.remoteFlag = 0
	}

	if q.remoteFlag == 0 {
		q.cond.Broadcast()
	}
}

func (q *Qos) Balance() {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	if q.remoteFlag == 1 {
		q.cond.Wait()
	}
}

func (q *Qos) Close() {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	q.remoteFlag = -1
	q.cond.Broadcast()
}

func NewQos(highWater, lowWater int, enterHighWaterCb, enterLowWaterCb func()) *Qos {
	if highWater <= lowWater {
		Panic("illegal qos indicator: %d - %d", highWater, lowWater)
	}

	var l sync.Mutex
	return &Qos{
		highWater:        highWater,
		lowWater:         lowWater,
		cond:             sync.NewCond(&l),
		enterHighWaterCb: enterHighWaterCb,
		enterLowWaterCb:  enterLowWaterCb}
}
