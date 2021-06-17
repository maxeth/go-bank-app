package config

import (
	"github.com/spf13/viper"
)

const (
	Source = "postgresql://postgres:secret@localhost:5432/bank_app?sslmode=disable"
	Driver = "postgres"
)

type Config struct {
	Environment   string
	DBString      string `mapstructure:"DB_STRING"`
	DBDriver      string `mapstructure:"DB_DRIVER"`
	ServerAddress string `mapstructure:"SERVER_ADDRESS"`
}

func New() (config Config, err error) {
	viper.AddConfigPath(".")
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.AutomaticEnv()
	err = viper.ReadInConfig()
	if err != nil {
		return Config{}, err
	}

	err = viper.Unmarshal(&config)

	return
}
