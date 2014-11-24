//
//   date  : 2014-06-04
//   author: xjdrew
//

package tunnel

func Recover() {
	if err := recover(); err != nil {
		LogStack("goroutine failed:%v", err)
	}
}
