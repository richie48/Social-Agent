package internal

import (
	"log"
	"os"
	"strings"
)

// LogLevel represents the logging level.
type LogLevel int

const (
	DebugLevel LogLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
)

// Logger provides structured logging with levels.
type Logger struct {
	level    LogLevel
	debugLog *log.Logger
	infoLog  *log.Logger
	warnLog  *log.Logger
	errorLog *log.Logger
}

// NewLogger creates a new logger with the specified level.
func NewLogger(logLevel string) *Logger {
	var level LogLevel
	switch strings.ToLower(logLevel) {
	case "debug":
		level = DebugLevel
	case "info":
		level = InfoLevel
	case "warn":
		level = WarnLevel
	case "error":
		level = ErrorLevel
	default:
		level = InfoLevel
	}

	return &Logger{
		level:    level,
		debugLog: log.New(os.Stdout, "[DEBUG] ", log.LstdFlags),
		infoLog:  log.New(os.Stdout, "[INFO] ", log.LstdFlags),
		warnLog:  log.New(os.Stdout, "[WARN] ", log.LstdFlags),
		errorLog: log.New(os.Stderr, "[ERROR] ", log.LstdFlags),
	}
}

// Debug logs a debug message.
func (l *Logger) Debug(format string, args ...interface{}) {
	if l.level <= DebugLevel {
		l.debugLog.Printf(format, args...)
	}
}

// Info logs an info message.
func (l *Logger) Info(format string, args ...interface{}) {
	if l.level <= InfoLevel {
		l.infoLog.Printf(format, args...)
	}
}

// Warn logs a warning message.
func (l *Logger) Warn(format string, args ...interface{}) {
	if l.level <= WarnLevel {
		l.warnLog.Printf(format, args...)
	}
}

// Error logs an error message.
func (l *Logger) Error(format string, args ...interface{}) {
	if l.level <= ErrorLevel {
		l.errorLog.Printf(format, args...)
	}
}
