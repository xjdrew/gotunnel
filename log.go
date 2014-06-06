//
//   date  : 2014-06-07
//   author: xjdrew
//

package main

func _print(level string, format string, a ...interface{}) {
	logger.Printf(level+format, a...)
}

func Debug(format string, a ...interface{}) {
	_print("<debug>", format, a...)
}

func Info(format string, a ...interface{}) {
	_print("<info>", format, a...)
}

func Error(format string, a ...interface{}) {
	_print("<error>", format, a...)
}

func Panic(format string, a ...interface{}) {
	_print("<panic>", format, a...)
  panic("!!")
}
