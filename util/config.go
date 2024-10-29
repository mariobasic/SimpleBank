package util

import (
	"github.com/spf13/viper"
	"strings"
	"time"
)

type Config struct {
	DB struct {
		Driver string `mapstructure:"driver"`
		Source string `mapstructure:"source"`
	} `mapstructure:"db"`
	Server struct {
		Address string `mapstructure:"address"`
	} `mapstructure:"server"`
	Token struct {
		SymmetricKey   string        `mapstructure:"symmetric_key"`
		AccessDuration time.Duration `mapstructure:"access_duration"`
	} `mapstructure:"token"`
}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("yml")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}
