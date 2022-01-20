package log

import (
	"fmt"
	"log"
	"os"
	"runtime/debug"
)

var infoLogger = log.New(os.Stdout, "[INFO] ", log.Ldate|log.Ltime|log.Lshortfile)
var warnLogger = log.New(os.Stdout, "[WARNING] ", log.Ldate|log.Ltime|log.Lshortfile)
var errorLogger = log.New(os.Stderr, "[ERROR] ", log.Ldate|log.Ltime|log.Lshortfile)
var fatalLogger = log.New(os.Stderr, "[FATAL] ", log.Ldate|log.Ltime|log.Lshortfile)

var Debug = false

func Info(args ...interface{}) {
	infoLogger.Output(2, fmt.Sprint(args...))
}

func Warning(args ...interface{}) {
	warnLogger.Output(2, fmt.Sprint(args...))
}

func Error(args ...interface{}) {
	errorLogger.Output(2, fmt.Sprint(args...))
	if Debug {
		errorLogger.Println(string(debug.Stack()))
	}
}

func Fatal(args ...interface{}) {
	fatalLogger.Output(2, fmt.Sprint(args...))
	errorLogger.Println(string(debug.Stack()))
	os.Exit(1)
}
