package log

import (
	"fmt"
	"log"
	"os"
)

var errorLogger = log.New(os.Stderr, "ERROR ", log.LstdFlags|log.Lmsgprefix)
var warnLogger = log.New(os.Stdout, "WARN ", log.LstdFlags|log.Lmsgprefix)
var enableDebug = false

func SetDebug(state bool) {
	enableDebug = state
}

func Print(v ...interface{}) {
	if !enableDebug {
		_ = log.Output(2, fmt.Sprint(v...))
	}
}

func Printf(format string, v ...interface{}) {
	if !enableDebug {
		_ = log.Output(2, fmt.Sprintf(format, v...))
	}
}

func Error(v ...interface{}) {
	_ = errorLogger.Output(2, fmt.Sprint(v...))
}

//noinspection GoUnusedExportedFunction
func Errorf(format string, a ...interface{}) {
	_ = errorLogger.Output(2, fmt.Sprintf(format, a...))
}

//noinspection GoUnusedExportedFunction
func Warn(v ...interface{}) {
	_ = warnLogger.Output(2, fmt.Sprint(v...))
}

//noinspection GoUnusedExportedFunction
func Warnf(format string, a ...interface{}) {
	_ = warnLogger.Output(2, fmt.Sprintf(format, a...))
}
