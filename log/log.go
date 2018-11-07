package log

import (
	stdlog "log"
	"os"
)

var debugFile, _ = os.OpenFile("debug.log", os.O_CREATE|os.O_APPEND, 0666)

var debug = stdlog.New(debugFile, "", stdlog.LstdFlags)
var stdout = stdlog.New(os.Stdout, "", stdlog.LstdFlags)

func DebugHttp(data []byte, err error) {
	if err == nil {
		debug.Printf("%s\n\n", data)
	} else {
		debug.Printf("%s\n\n", err)
	}
}

func Debugf(format string, v ...interface{}) {
	debug.Printf(format+"\n", v...)
}

func Errorf(format string, v ...interface{}) {
	debug.Printf(format+"\n", v...)
}

func Infof(format string, v ...interface{}) {
	stdout.Printf(format+"\n", v...)
}

func Infoln(msg string) {
	stdout.Printf(msg)
}

func Info(v ...interface{}) {
	stdout.Print(v...)
}
