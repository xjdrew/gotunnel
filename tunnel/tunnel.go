//
//   date  : 2014-06-04
//   author: xjdrew
//

package tunnel

import (
	"time"
)

const (
	TunnelMaxId           = ^uint16(0)
	TunnelMinSpan         = 3 // 3次心跳无回应则断开
	TunnelPacketSize      = 8192
	TunnelKeepAlivePeriod = time.Second * 180
)

var (
	Heartbeat int  = 1
	Timeout   int  = 0
	LogLevel  uint = 1
	mpool          = NewMPool(TunnelPacketSize)
)
