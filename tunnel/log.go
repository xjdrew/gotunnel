//
//   date  : 2014-06-07
//   author: xjdrew
//

package tunnel

import (
	"io"
	"log"
	"os"
	"runtime"
)

var logger *log.Logger

func init() {
	//logger = log.New(io.Writer(os.Stderr), "", log.Ldate | log.Lmicroseconds | log.Lshortfile)
	logger = log.New(io.Writer(os.Stderr), "", log.Ldate|log.Lmicroseconds)
}

func _print(format string, a ...interface{}) {
	logger.Printf(format, a...)
}

func Trace(format string, a ...interface{}) {
	if LogLevel > 3 {
		_print(format, a...)
	}
}

func Debug(format string, a ...interface{}) {
	if LogLevel > 2 {
		_print(format, a...)
	}
}

func Info(format string, a ...interface{}) {
	if LogLevel > 1 {
		_print(format, a...)
	}
}

func Error(format string, a ...interface{}) {
	if LogLevel > 0 {
		_print(format, a...)
	}
}

func Log(format string, a ...interface{}) {
	_print(format, a...)
}

func LogStack(format string, a ...interface{}) {
	_print(format, a...)

	buf := make([]byte, 32768)
	runtime.Stack(buf, true)
	_print("!!!!!stack!!!!!: %s", buf)
}

func LogCurStack(format string, a ...interface{}) {
	_print(format, a...)
	buf := make([]byte, 8192)
	runtime.Stack(buf, false)
	_print("!!!!!stack!!!!!: %s", buf)
}

func Panic(format string, a ...interface{}) {
	LogStack(format, a...)
	panic("!!")
}
