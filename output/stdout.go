package output

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/chihqiang/dbxgo/types"
)

// StdoutOutput Console output implementation
type StdoutOutput struct{}

// NewStdoutOutput Creates a StdoutOutput instance
func NewStdoutOutput() (*StdoutOutput, error) {
	return &StdoutOutput{}, nil
}

// Send Outputs the event to the console
func (s *StdoutOutput) Send(ctx context.Context, event types.EventData) error {
	data, err := json.MarshalIndent(event, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

// Close No resources to close for console output
func (s *StdoutOutput) Close() error {
	return nil
}
