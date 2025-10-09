package config

import (
	"fmt"
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/chihqiang/dbxgo/output"
	"github.com/chihqiang/dbxgo/source"
	"github.com/chihqiang/dbxgo/store"
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
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

// Load 尝试加载配置。
// 加载顺序：优先读取文件 → 若文件不存在或解析失败，则从环境变量加载。
func Load(path string) (*Config, error) {
	var cfg Config

	// ① 优先尝试从配置文件加载
	data, err := os.ReadFile(path)
	if err == nil {
		// 尝试解析 YAML 配置文件
		if yamlErr := yaml.Unmarshal(data, &cfg); yamlErr == nil {
			return &cfg, nil // 文件读取并解析成功
		}
		// 如果 YAML 解析失败，则继续尝试环境变量
	}

	// ② 文件加载失败，尝试从环境变量加载
	if envErr := env.Parse(&cfg); envErr == nil {
		return &cfg, nil // 环境变量解析成功
	}

	// ③ 若两种方式都失败，返回错误信息
	return nil, fmt.Errorf("failed to load configuration (file: %v, env: %v)", err, env.Parse(&cfg))
}
