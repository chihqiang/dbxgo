package logx

import (
	"io"
	"os"
	"sync"
)

var (
	std     *Logger   // Global default Logger instance
	stdOnce sync.Once // Ensures the global Logger is initialized only once (thread-safe)
)

// _std returns the global Logger instance (singleton pattern)
// Initializes the Logger on first call and sets callDepth
func _std() *Logger {
	stdOnce.Do(func() {
		// Create a Logger that outputs to standard error
		_log := New(os.Stderr)
		_log.callDepth = 3 // Call depth offset for correct file/line number display
		std = _log
	})
	return std
}

// SetOutput sets the output destination for the global Logger (thread-safe)
func SetOutput(w io.Writer) {
	_std().SetOutput(w)
}

// SetPrefix sets the log prefix for the global Logger (thread-safe)
func SetPrefix(p string) {
	_std().SetPrefix(p)
}

// SetFormatter sets the log formatting function for the global Logger (thread-safe)
func SetFormatter(fn Formatter) {
	_std().SetFormatter(fn)
}

// Debug logs a message at Debug level
func Debug(format string, v ...any) {
	_std().Debug(format, v...)
}

// Info logs a message at Info level
func Info(format string, v ...any) {
	_std().Info(format, v...)
}

// Warn logs a message at Warn level
func Warn(format string, v ...any) {
	_std().Warn(format, v...)
}

// Error logs a message at Error level
func Error(format string, v ...any) {
	_std().Error(format, v...)
}

// Log records a log message at the specified Level
// The log will be output if the level is higher than the Logger's set minimum level
func Log(level Level, format string, v ...any) error {
	return _std().Log(level, format, v...)
}
