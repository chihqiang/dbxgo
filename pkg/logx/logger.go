package logx

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
	"time"
)

// ILogger defines the interface for log operations
type ILogger interface {
	SetOutput(w io.Writer)
	SetPrefix(prefix string)
	SetFormatter(fn Formatter)
	Debug(format string, v ...any)
	Info(format string, v ...any)
	Warn(format string, v ...any)
	Error(format string, v ...any)
	Log(level Level, format string, v ...any) error
}

// New creates a new Logger instance
// Parameter w specifies the log output destination (can be os.Stdout, os.Stderr, file, etc.)
func New(w io.Writer) *Logger {
	l := &Logger{}
	l.SetOutput(w)
	l.SetFormatter(DefaultFormatter) // Use default formatting function
	return l
}

// Logger represents a logging object
type Logger struct {
	mu        sync.RWMutex // Read-write lock for concurrent safety
	writer    io.Writer    // Log output destination
	prefix    string       // Log prefix
	formatter Formatter    // Log formatting function
	callDepth int          // Offset for runtime.Caller to correctly display caller file and line number
}

// SetOutput sets the log output destination (thread-safe)
func (l *Logger) SetOutput(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.writer = w
}

// SetPrefix sets the log prefix (thread-safe)
func (l *Logger) SetPrefix(prefix string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.prefix = prefix
}

// SetFormatter sets the log formatting function (thread-safe)
func (l *Logger) SetFormatter(fn Formatter) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.formatter = fn
}

// Debug outputs a log message at Debug level
func (l *Logger) Debug(format string, v ...any) {
	_ = l.log(LevelDebug, format, v...)
}

// Info outputs a log message at Info level
func (l *Logger) Info(format string, v ...any) {
	_ = l.log(LevelInfo, format, v...)
}

// Warn outputs a log message at Warn level
func (l *Logger) Warn(format string, v ...any) {
	_ = l.log(LevelWarn, format, v...)
}

// Error outputs a log message at Error level
func (l *Logger) Error(format string, v ...any) {
	_ = l.log(LevelError, format, v...)
}

func (l *Logger) Log(level Level, format string, v ...any) error {
	return l.log(level, format, v...)
}

// log outputs a log message at the specified level
// 1. Gets caller file and line number based on callDepth
// 2. Formats the log entry using the Formatter
// 3. Writes to the log output destination (writer), defaults to os.Stdout if writer is nil
func (l *Logger) log(level Level, format string, v ...any) error {
	// Concurrent-safe read of current Logger state
	l.mu.RLock()
	prefix := l.prefix
	formatter := l.formatter
	writer := l.writer
	callDepth := l.callDepth
	if callDepth == 0 {
		callDepth = 2
	}
	l.mu.RUnlock()

	// Format log content
	msg := fmt.Sprintf(format, v...)

	// Get caller file and line number
	_, file, line, ok := runtime.Caller(callDepth)
	if !ok {
		file = "???" // Use placeholder when unable to retrieve
		line = 0
	}
	// Output log, default to stdout if writer is nil
	if writer == nil {
		writer = os.Stdout
	}
	_, err := writer.Write(formatter(LogEntry{
		Time:      time.Now(),
		Level:     level,
		Prefix:    prefix,
		CallDepth: callDepth,
		File:      file,
		Line:      line,
		Message:   msg,
	}))
	return err
}
