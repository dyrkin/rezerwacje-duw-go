package log

import (
	stdlog "log"
	"os"
)

var debugFile, _ = os.OpenFile("debug.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

var debug = stdlog.New(debugFile, "", stdlog.LstdFlags)
var stdout = stdlog.New(os.Stdout, "", stdlog.LstdFlags)

func Debugf(format string, v ...interface{}) {
	debug.Printf(format+"\n", v...)
}

func Errorf(format string, v ...interface{}) {
	debug.Printf(format+"\n", v...)
}

func Infof(format string, v ...interface{}) {
	Debugf(format, v...)
	stdout.Printf(format+"\n", v...)
}
