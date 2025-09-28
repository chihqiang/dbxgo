package config

import (
	"github.com/chihqiang/dbxgo/output"
	"github.com/chihqiang/dbxgo/source"
	"github.com/chihqiang/dbxgo/store"
	"gopkg.in/yaml.v3"
	"os"
)

// Config 定义全局配置结构
// 用于从配置文件加载应用程序的所有配置项

type Config struct {
	Store  store.Config  `yaml:"store"`
	Source source.Config `yaml:"source"`
	Output output.Config `yaml:"output"`
}

// Load 从指定路径加载YAML配置文件
// path: 配置文件的路径
// 返回值: 配置对象指针和可能的错误
func Load(path string) (*Config, error) {
	// 读取配置文件内容
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// 将YAML数据解析到Config结构体
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
