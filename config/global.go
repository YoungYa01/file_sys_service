package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"os"
)

type GlobalConfig struct {
	App      AppConfig      `yaml:"app"`
	Jwt      JwtConfig      `yaml:"jwt"`
	Database DatabaseConfig `yaml:"database"`
}

type AppConfig struct {
	Name     string `yaml:"name"`
	Version  string `yaml:"version"`
	Env      string `yaml:"env"`
	Port     int    `yaml:"port"`
	LogLevel string `yaml:"log_level"`
}

type JwtConfig struct {
	Secret string `yaml:"secret"`
	Expire int    `yaml:"expire"`
	Issuer string `yaml:"issuer"`
}

type DatabaseConfig struct {
	URL             string `yaml:"url"`
	UserName        string `yaml:"username"`
	Password        string `yaml:"password"`
	DriverClassName string `yaml:"driverClassName"`
	Dialect         string `yaml:"dialect"`
}

var AppGlobalConfig = new(GlobalConfig)

var DB *gorm.DB

func LoadConfig() error {
	// 加载配置文件
	file, err := os.ReadFile("./config/config.yaml")
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %v", err)
	}

	if err := yaml.Unmarshal(file, &AppGlobalConfig); err != nil {
		return fmt.Errorf("解析配置文件失败: %v", err)
	}

	DB, err = gorm.Open(mysql.Open(AppGlobalConfig.Database.URL), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("数据库连接失败: %v", err)
	}

	return nil
}
