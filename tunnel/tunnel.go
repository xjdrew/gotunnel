//
//   date  : 2014-06-04
//   author: xjdrew
//

package tunnel

import (
	"time"
)

const (
	TunnelMaxId = ^uint16(0)

	TunnelPacketSize      = 8192
	TunnelKeepAlivePeriod = time.Second * 180

	defaultHeartbeat = 1
	tunnelMinSpan    = 3 // 3次心跳无回应则断开
)

var (
	// Heartbeat interval for tunnel heartbeat, seconds.
	Heartbeat int = 1 // seconds

	// Timeout for tunnel write/read, seconds
	Timeout int = 0 //

	// LogLevel .
	LogLevel uint = 1
	mpool         = NewMPool(TunnelPacketSize)
)

func getHeartbeat() time.Duration {
	if Heartbeat <= 0 {
		Heartbeat = defaultHeartbeat
	}
	return time.Duration(Heartbeat) * time.Second
}

func getTimeout() time.Duration {
	return time.Duration(Timeout) * time.Second
}
