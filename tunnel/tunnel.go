//
//   date  : 2014-06-04
//   author: xjdrew
//

package tunnel

// tunnel read/write timeout
const (
	PacketSize = 8192
)

var (
	Timeout  int64 = 0
	LogLevel uint  = 1
	mpool          = NewMPool(PacketSize)
)

type Service interface {
	Start() error
	Status()
}
