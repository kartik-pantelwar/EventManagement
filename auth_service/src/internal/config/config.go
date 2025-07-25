package config

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	DB_USER    string `mapstructure:"DB_USER"`
	DB_HOST    string `mapstructure:"DB_HOST"`
	DB_PORT    string `mapstructure:"DB_PORT"`
	DB_PASS    string `mapstructure:"DB_PASS"`
	DB_NAME    string `mapstructure:"DB_NAME"`
	DB_SSLMODE string `mapstructure:"DB_SSLMODE"`
	APP_ENV    string `mapstructure:"APP_ENV"`
	APP_PORT   string `mapstructure:"APP_PORT"`
}

func Loadconfig() (*Config, error) {
	config := &Config{}
	env := "local"
	envConfigFileName := fmt.Sprintf(".env.%s", env)

	viper.AutomaticEnv()
	viper.AddConfigPath("./.secrets")
	viper.SetConfigName(envConfigFileName)
	viper.SetConfigType("env")

	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("Config File not found using environment variables")
		} else {
			return nil, fmt.Errorf("Failed to read config File :%w", err)
		}
	}

	err = viper.Unmarshal(&config)
	if err != nil {
		return nil, fmt.Errorf("Failed to Unmarshal config :%w", err)
	}
	return config, nil
}
