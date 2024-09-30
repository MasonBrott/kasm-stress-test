package utils

import (
	"log"
	"os"
	"sync"
)

var (
	infoLogger  *log.Logger
	errorLogger *log.Logger
	once        sync.Once
)

// InitLoggers initializes the loggers
func InitLoggers() {
	once.Do(func() {
		infoLogger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
		errorLogger = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	})
}

// Info logs an info message
func Info(format string, v ...interface{}) {
	infoLogger.Printf(format, v...)
}

// Error logs an error message
func Error(format string, v ...interface{}) {
	errorLogger.Printf(format, v...)
}

// Fatal logs an error message and then exits the program
func Fatal(format string, v ...interface{}) {
	errorLogger.Fatalf(format, v...)
}
