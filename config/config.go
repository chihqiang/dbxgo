package config

import (
	"github.com/caarlos0/env/v11"
	"github.com/chihqiang/dbxgo/output"
	"github.com/chihqiang/dbxgo/source"
	"github.com/chihqiang/dbxgo/store"
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
	"os"
)

func init() {
	_ = godotenv.Load()
}

// Config 定义全局配置结构
// 用于从配置文件加载应用程序的所有配置项
type Config struct {
	Store  store.Config  `yaml:"store" json:"store" mapstructure:"store"`
	Source source.Config `yaml:"source" json:"source" mapstructure:"source"`
	Output output.Config `yaml:"output" json:"output" mapstructure:"output"`
}

func LoadEnv() (*Config, error) {
	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Load 从指定路径加载配置文件（YAML），并支持环境变量优先覆盖
// path: 配置文件路径
// 返回值: 配置对象指针和可能的错误
func Load(path string) (*Config, error) {
	// 尝试从环境变量中加载配置（优先级最高）
	conf, err := LoadEnv()
	if err == nil && conf != nil {
		return conf, nil // 环境变量加载成功则直接返回
	}
	// 读取指定路径的配置文件内容
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err // 文件读取失败返回错误
	}
	var cfg Config
	yamlErr := yaml.Unmarshal(data, &cfg)
	if yamlErr != nil {
		return nil, yamlErr
	}
	// 返回解析后的配置对象指针
	return &cfg, nil
}
