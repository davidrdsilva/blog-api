package logging

import (
	"fmt"
	"time"
)

// Color codes for terminal output
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
)

// LogLevel represents the severity of a log message
type LogLevel string

const (
	INFO  LogLevel = "INFO"
	WARN  LogLevel = "WARN"
	ERROR LogLevel = "ERROR"
	DEBUG LogLevel = "DEBUG"
)

// Logger provides structured, colorful logging
type Logger struct {
	serviceName string
}

// NewLogger creates a new logger instance
func NewLogger(serviceName string) *Logger {
	return &Logger{
		serviceName: serviceName,
	}
}

// Info logs an informational message in green
func (l *Logger) Info(message string, fields ...Field) {
	l.log(INFO, ColorGreen, message, fields...)
}

// Warn logs a warning message in yellow
func (l *Logger) Warn(message string, fields ...Field) {
	l.log(WARN, ColorYellow, message, fields...)
}

// Error logs an error message in red
func (l *Logger) Error(message string, fields ...Field) {
	l.log(ERROR, ColorRed, message, fields...)
}

// Debug logs a debug message in cyan
func (l *Logger) Debug(message string, fields ...Field) {
	l.log(DEBUG, ColorCyan, message, fields...)
}

// log is the internal logging function
func (l *Logger) log(level LogLevel, color string, message string, fields ...Field) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	// Format: [timestamp] [LEVEL] message | key=value key=value
	logLine := fmt.Sprintf("%s[%s] [%s]%s %s",
		color,
		timestamp,
		level,
		ColorReset,
		message,
	)

	if len(fields) > 0 {
		logLine += " |"
		for _, field := range fields {
			logLine += fmt.Sprintf(" %s=%v", field.Key, field.Value)
		}
	}

	fmt.Println(logLine)
}

// Field represents a key-value pair for structured logging
type Field struct {
	Key   string
	Value interface{}
}

// F is a helper function to create a Field
func F(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}
