package util

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	DBDriver            string        `mapstructure:"DB_DRIVER"`
	DBSource            string        `mapstructure:"DB_SOURCE"`
	ServerAddress       string        `mapstructure:"SERVER_ADDRESS"`
	TokenSymmetricKey   string        `mapstructure:"TOKEN_SYMMETRIC_KEY"`
	AccessTokenDuration time.Duration `mapstructure:"ACCESS_TOKEN_DURATION"`
	RedisAddr           string        `mapstructure:"REDIS_ADDR"`
	RedisPassword       string        `mapstructure:"REDIS_PASSWORD"`
	EmailSenderName     string        `mapstructure:"EMAIL_SENDER_NAME"`
	EmailSenderAddr     string        `mapstructure:"EMAIL_SENDER_ADDR"`
	EmailPassword       string        `mapstructure:"EMAIL_PASSWORD"`
}

func LoadConfig(path string) (config Config, err error) {
	v := viper.New()
	v.AddConfigPath(path)
	v.SetConfigName("app")
	v.SetConfigType("env")
	v.AutomaticEnv()

	for _, key := range []string{
		"DB_DRIVER",
		"DB_SOURCE",
		"SERVER_ADDRESS",
		"TOKEN_SYMMETRIC_KEY",
		"ACCESS_TOKEN_DURATION",
		"REDIS_ADDR",
		"REDIS_PASSWORD",
		"EMAIL_SENDER_NAME",
		"EMAIL_SENDER_ADDR",
		"EMAIL_PASSWORD",
	} {
		if err := v.BindEnv(key); err != nil {
			return config, err
		}
	}

	err = v.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return config, err
		}
	}

	err = v.Unmarshal(&config)
	if err != nil {
		return config, err
	}

	return
}
