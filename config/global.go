package config

import (
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

func LoadConfig() {
	// 加载配置文件
	file, err := os.ReadFile("./config/config.yaml")
	if err != nil {
		return
	}
	err = yaml.Unmarshal(file, &AppGlobalConfig)

	db, err := gorm.Open(mysql.Open(AppGlobalConfig.Database.URL), &gorm.Config{})
	if err != nil {
		return
	}
	DB = db
}
