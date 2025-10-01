package source

import (
	"context"
	"fmt"
	"github.com/chihqiang/dbxgo/store"
	"github.com/chihqiang/dbxgo/types"
)

const (
	StoreKeyPosition = "_dbxgo_position"
)

type SourceType string

var (
	SourceTypeMysql SourceType = "mysql"
)

// Config 定义数据源配置结构
// 用于配置数据库连接信息和存储配置
type Config struct {
	// Type 数据源类型，如 "mysql"
	Type  SourceType   `yaml:"type" json:"type" mapstructure:"type" env:"SOURCE_TYPE,required"`
	Mysql MysqlConfig  `yaml:"mysql" json:"mysql" mapstructure:"mysql"`
	Store store.Config `yaml:"store" json:"store" mapstructure:"store"`
}

// ISource 定义数据源接口
// 所有具体数据源实现必须实现此接口
type ISource interface {
	// WithStore 设置存储接口，用于持久化或读取偏移量及状态
	WithStore(store store.IStore)

	// Run 启动数据源监听
	// ctx: 上下文，用于控制取消和超时
	// 返回值: 可能的错误
	Run(ctx context.Context) error

	// GetChanEventData 返回事件数据通道
	// 外部通过此通道接收数据库变更事件
	// 返回值: 只读的事件数据通道
	GetChanEventData() <-chan types.EventData

	// Close 关闭数据源，释放资源
	// 返回值: 可能的错误
	Close() error
}

// NewSource 根据配置创建对应的数据源实例
// cfg: 数据源配置信息
// 返回值: 数据源接口实现和可能的错误
func NewSource(cfg Config) (ISource, error) {
	switch cfg.Type {
	case SourceTypeMysql:
		return NewMySQLSource(cfg.Mysql)
	default:
		return nil, fmt.Errorf("unsupported source type: %s", cfg.Type)
	}
}
