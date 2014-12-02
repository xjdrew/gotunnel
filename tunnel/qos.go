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
	remoteFlag       bool       // 对端高水位标志
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
	cond := q.cond
	if cond == nil {
		return
	}

	cond.L.Lock()
	q.remoteFlag = flag
	if flag == false {
		cond.Broadcast()
	}
	cond.L.Unlock()
}

func (q *Qos) Balance() {
	cond := q.cond
	if cond == nil {
		return
	}

	cond.L.Lock()
	if q.remoteFlag {
		cond.Wait()
	}
	cond.L.Unlock()
}

func (q *Qos) Close() {
	cond := q.cond
	if cond != nil {
		q.cond = nil
		cond.Broadcast()
	}
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
