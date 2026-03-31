package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// ServerConfig 服务器配置
type ServerConfig struct {
	Port int `mapstructure:"port"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Driver string `mapstructure:"driver"`
	DSN    string `mapstructure:"dsn"`
}

// JWTConfig JWT 配置
type JWTConfig struct {
	Secret string `mapstructure:"secret"`
}

// WechatConfig 微信小程序及支付配置
type WechatConfig struct {
	AppID            string `mapstructure:"app_id"`
	Secret           string `mapstructure:"secret"`
	MchID            string `mapstructure:"mch_id"`
	MchAPIKey        string `mapstructure:"mch_api_key"`
	MchSerialNo      string `mapstructure:"mch_serial_no"`
	MchPrivateKeyPath string `mapstructure:"mch_private_key_path"`
	NotifyURL        string `mapstructure:"notify_url"`
}

// Config 全局配置
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
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
