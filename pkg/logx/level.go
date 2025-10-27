package logx

import (
	"fmt"
	"github.com/fatih/color"
)

type Level int

const (
	LevelDebug Level = -4 // Debug level
	LevelInfo  Level = 0  // General information
	LevelWarn  Level = 4  // Warning
	LevelError Level = 8  // Error
)

// String returns the string representation of the log level
// For non-standard levels, it appends the offset to the base level, e.g., DEBUG+1
func (l Level) String() string {
	str := func(base string, val Level) string {
		if val == 0 {
			return base
		}
		return fmt.Sprintf("%s%+d", base, val)
	}

	switch {
	case l < LevelInfo:
		return str("DEBUG", l-LevelDebug)
	case l < LevelWarn:
		return str("INFO", l-LevelInfo)
	case l < LevelError:
		return str("WARN", l-LevelWarn)
	default:
		return str("ERROR", l-LevelError)
	}
}

// MarshalJSON implements the json.Marshaler interface
// This method is called when LogEntry is serialized with json.Marshal
// The purpose is to serialize the Level type into its corresponding string (e.g., "INFO", "ERROR") instead of an integer
func (l Level) MarshalJSON() ([]byte, error) {
	// Call l.String() to get the string corresponding to Level
	// Then add double quotes and return as a byte slice to conform to JSON string format
	return []byte(`"` + l.String() + `"`), nil
}

// Color returns the color output corresponding to the log level (using github.com/fatih/color)
func (l Level) Color() *color.Color {
	switch {
	case l >= LevelError: // 8 and above
		return color.New(color.FgHiRed, color.Bold)
	case l >= LevelWarn: // 4 and above
		return color.New(color.FgYellow, color.Bold)
	case l >= LevelInfo: // 0 and above
		return color.New(color.FgGreen)
	case l >= LevelDebug: // -4 and above
		return color.New(color.FgBlue)
	default:
		return color.New(color.FgWhite)
	}
}
