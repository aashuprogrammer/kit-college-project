package utils

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	DBUrl               string        `mapstructure:"DATABASE_URL"`
	Port                int16         `mapstructure:"PORT"`
	TokenSymmectricKey  string        `mapstructure:"TOKEN_SYMMETRIC_KEY"`
	TokenDuration       time.Duration `mapstructure:"TOKEN_DURATION"`
	ProfilesFolder      string        `mapstructure:"PROFILES_FOLDER"`
	CashfreeAppID       string        `mapstructure:"CASHFREE_APP_ID"`
	CashfreeSecretKey   string        `mapstructure:"CASHFREE_SECRET_KEY"`
	CashfreeEnvironment string        `mapstructure:"CASHFREE_ENVIRONMENT"` // sandbox or production
}

func LoadConfig(path string) (Config, error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	var config Config
	err := viper.ReadInConfig()
	if err != nil {
		return config, err
	}
	err = viper.Unmarshal(&config)

	return config, err
}
