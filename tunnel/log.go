//
//   date  : 2014-06-07
//   author: xjdrew
//

package tunnel

import (
	"io"
	"log"
	"os"
)

var logger *log.Logger

func init() {
	//logger = log.New(io.Writer(os.Stderr), "", log.Ldate | log.Lmicroseconds | log.Lshortfile)
	logger = log.New(io.Writer(os.Stderr), "", log.Ldate|log.Lmicroseconds)
}

func _print(format string, a ...interface{}) {
	logger.Printf(format, a...)
}

func Debug(format string, a ...interface{}) {
	if options.LogLevel > 2 {
		_print(format, a...)
	}
}

func Info(format string, a ...interface{}) {
	if options.LogLevel > 1 {
		_print(format, a...)
	}
}

func Error(format string, a ...interface{}) {
	if options.LogLevel > 0 {
		_print(format, a...)
	}
}

func Panic(format string, a ...interface{}) {
	_print(format, a...)
	panic("!!")
}
