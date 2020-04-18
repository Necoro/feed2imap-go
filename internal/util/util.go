package util

import (
	"log"
	"os"
)

var errorLogger = log.New(os.Stderr, "ERROR ", log.LstdFlags|log.Lmsgprefix)

func Error(v ...interface{}) {
	errorLogger.Print(v...)
}

//noinspection GoUnusedExportedFunction
func Errorf(format string, a ...interface{}) {
	errorLogger.Printf(format, a...)
}
