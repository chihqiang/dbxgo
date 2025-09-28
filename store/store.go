package store

import "fmt"

type StoreType string

const (
	// FileStoreType 类型
	FileStoreType StoreType = "file"
	// RedisStoreType 类型
	RedisStoreType StoreType = "redis"
)

var (
	stores = map[StoreType]func(cfg Config) (IStore, error){
		FileStoreType: func(cfg Config) (IStore, error) {
			return NewFileStore(cfg.File)
		},
		RedisStoreType: func(cfg Config) (IStore, error) {
			return NewRedisStore(cfg.Redis)
		},
	}
)

// Config 存储配置结构
type Config struct {
	// Type 存储类型，例如 "file"
	Type  StoreType   `yaml:"type"`
	File  FileConfig  `yaml:"file"`
	Redis RedisConfig `yaml:"redis"`
}

type IStore interface {
	Has(key string) bool
	Set(key string, value []byte) error
	Get(key string) ([]byte, error)
}

func NewStore(cfg Config) (IStore, error) {
	// 查找对应的构造函数
	creator, exists := stores[cfg.Type]
	if !exists {
		return nil, fmt.Errorf("store type %s is not registered", cfg.Type)
	}
	// 调用构造函数创建输出实例
	return creator(cfg)
}
