package util

import (
	"github.com/spf13/viper"
	"strings"
	"time"
)

type Config struct {
	Env string `mapstructure:"env"`
	DB  struct {
		Driver       string `mapstructure:"driver"`
		Source       string `mapstructure:"source"`
		MigrationURL string `mapstructure:"migration_url"`
	} `mapstructure:"db"`
	Server struct {
		Http  string `mapstructure:"http_address"`
		Grpc  string `mapstructure:"grpc_address"`
		Redis string `mapstructure:"redis_address"`
	} `mapstructure:"server"`
	Token struct {
		SymmetricKey    string        `mapstructure:"symmetric_key"`
		AccessDuration  time.Duration `mapstructure:"access_duration"`
		RefreshDuration time.Duration `mapstructure:"refresh_duration"`
	} `mapstructure:"token"`
	Email struct {
		Sender struct {
			Name     string `mapstructure:"name"`
			Address  string `mapstructure:"address"`
			Password string `mapstructure:"password"`
		} `mapstructure:"sender"`
	} `mapstructure:"email"`
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
