package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// ServerConfig 定义服务器相关配置
// DatabaseConfig 定义数据库相关配置
// Config 聚合所有配置项

type ServerConfig struct {
	Port int `mapstructure:"port"`
}

type DatabaseConfig struct {
	Driver string `mapstructure:"driver"`
	DSN    string `mapstructure:"dsn"`
}

// AdminConfig 定义后台管理员账户配置
type AdminConfig struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

// JWTConfig 定义 JWT 相关配置
type JWTConfig struct {
	Secret string `mapstructure:"secret"`
}

// WechatConfig 定义微信小程序相关配置
type WechatConfig struct {
	AppID  string `mapstructure:"app_id"`
	Secret string `mapstructure:"secret"`
}

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Admin    AdminConfig    `mapstructure:"admin"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Wechat   WechatConfig   `mapstructure:"wechat"`
}

// LoadConfig 从配置文件和环境变量加载配置
func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	exePath, err := os.Executable()
	if err != nil {
		return nil, err
	}
	exeDir := filepath.Dir(exePath)
	viper.AddConfigPath(filepath.Join(exeDir, "configs"))
	viper.AddConfigPath("configs")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
