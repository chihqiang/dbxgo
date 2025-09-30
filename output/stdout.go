package output

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/chihqiang/dbxgo/types"
)

// StdoutOutput 控制台输出实现
type StdoutOutput struct{}

// NewStdoutOutput 创建 StdoutOutput 实例
func NewStdoutOutput() (*StdoutOutput, error) {
	return &StdoutOutput{}, nil
}

// Send 输出事件到控制台
func (s *StdoutOutput) Send(ctx context.Context, event types.EventData) error {
	data, err := json.MarshalIndent(event, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

// Close 控制台输出无需关闭资源
func (s *StdoutOutput) Close() error {
	return nil
}
