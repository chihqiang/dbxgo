package logx

import (
	"fmt"
	"github.com/fatih/color"
	"strings"
	"time"
)

// LogEntry represents a log entry structure
type LogEntry struct {
	Time      time.Time `json:"time" xml:"time"`       // The time when the log occurred
	Level     Level     `json:"level" xml:"level"`     // Log level (e.g., TRACE, INFO, ERROR, etc.)
	Prefix    string    `json:"prefix" xml:"prefix"`   // Log prefix, used to distinguish modules or subsystems, can be empty
	File      string    `json:"file" xml:"file"`       // Path of the file where the log is located (relative path or formatted path)
	Line      int       `json:"line" xml:"line"`       // Line number in the file where the log is located
	Message   string    `json:"message" xml:"message"` // Main content of the log
	CallDepth int       `json:"-" xml:"-"`             // Stack depth, used to get the caller's position (file and line number)
}

// Formatter defines the type for formatting functions
// Input: log entry
// Output: formatted log string
type Formatter func(entry LogEntry) []byte

var DefaultFormatter Formatter = func(entry LogEntry) []byte {
	// Time format
	timestamp := entry.Time.Format("2006-01-02 15:04:05")
	// Log level in uppercase
	level := entry.Level.String()

	fileLine := fmt.Sprintf("[%s:%d]", TrimCallerPath(entry.File, 1), entry.Line)
	// Log prefix
	prefix := ""
	if entry.Prefix != "" {
		prefix = entry.Prefix + ": "
	}
	// Custom default output format
	logStr := fmt.Sprintf("%s %s %s %s%s",
		timestamp,
		entry.Level.Color().Sprint(level),
		color.New(color.FgHiBlack).Sprint(fileLine),
		color.New(color.FgHiBlack).Add(color.Bold).Sprint(prefix),
		entry.Level.Color().Sprint(entry.Message),
	)
	return []byte(logStr + "\n")
}

func TrimCallerPath(path string, n int) string {
	// Lovely borrowed from zap
	// Note: To ensure we trim the path correctly on Windows too, we
	// counter-intuitively need to use '/' and *not* os.PathSeparator here,
	// because the path comes from Go's standard library, specifically
	// runtime.Caller() which (as of March 2017) returns forward slashes even on
	// Windows.
	//
	// See https://github.com/golang/go/issues/3335
	// and https://github.com/golang/go/issues/18151
	//
	// for discussion on the issue in the Go project.
	// Return the full path if n is 0.
	if n <= 0 {
		return path
	}
	// Find the last separator.
	idx := strings.LastIndexByte(path, '/')
	if idx == -1 {
		return path
	}
	for i := 0; i < n-1; i++ {
		// Find the penultimate separator.
		idx = strings.LastIndexByte(path[:idx], '/')
		if idx == -1 {
			return path
		}
	}
	return path[idx+1:]
}
