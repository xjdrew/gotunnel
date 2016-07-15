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
	TunnelMaxTimeout      = 3600
	TunnelPacketSize      = 8192
	TunnelKeepAlivePeriod = time.Second * 180
)

var (
	Timeout  int  = 0
	LogLevel uint = 1
	Udt      bool = false
	mpool         = NewMPool(TunnelPacketSize)
)
