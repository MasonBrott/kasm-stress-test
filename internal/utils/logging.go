package utils

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
)

var (
	fileLogger    *log.Logger
	consoleLogger *log.Logger
	logFile       *os.File
	once          sync.Once
)

// InitLoggers initializes the loggers
func InitLoggers() error {
	var initError error
	once.Do(func() {
		// Create log file in the same directory as the executable
		execPath, err := os.Executable()
		if err != nil {
			initError = fmt.Errorf("error getting executable path: %w", err)
			return
		}
		logFilePath := filepath.Join(filepath.Dir(execPath), "stress-test.log")
		logFile, err = os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			initError = fmt.Errorf("error opening log file: %w", err)
			return
		}

		fileLogger = log.New(logFile, "", log.Ldate|log.Ltime|log.Lshortfile)
		consoleLogger = log.New(os.Stdout, "", 0) // Minimal console output
	})
	return initError
}

// CloseLogFile closes the log file
func CloseLogFile() {
	if logFile != nil {
		logFile.Close()
	}
}

// Info logs an info message to file and prints a simplified version to console
func Info(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	fileLogger.Printf("INFO: %s", message)
}

// Console prints a message to the console
func Console(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	consoleLogger.Print(message)
}

// Error logs an error message to file and console
func Error(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	fileLogger.Printf("ERROR: %s", message)
	consoleLogger.Printf("ERROR: %s", message)
	consoleLogger.Println("See stress-test.log for more info")
}

// Fatal logs an error message and then exits the program
func Fatal(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	fileLogger.Printf("FATAL: %s", message)
	consoleLogger.Printf("FATAL: %s", message)
	consoleLogger.Println("See stress-test.log for more info")
	os.Exit(1)
}
