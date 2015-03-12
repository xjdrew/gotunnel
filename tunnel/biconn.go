//
//   date  : 2015-03-10
//   author: xjdrew
//

package tunnel

import (
	"net"
)

type BiConn interface {
	net.Conn
	CloseRead() error
	CloseWrite() error
}
