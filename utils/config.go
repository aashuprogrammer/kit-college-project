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
	CashfreeAppID       string        `mapstructure:"CASHFREE_CLIENT_ID"`
	CashfreeSecretKey   string        `mapstructure:"CASHFREE_CLIENT_SECRET"`
	CashfreeEnvironment string        `mapstructure:"CASHFREE_ENV"` // sandbox or production
	R2AccountID         string        `mapstructure:"R2_ACCOUNT_ID"`
	R2AccessKeyID       string        `mapstructure:"R2_ACCESS_KEY_ID"`
	R2SecretAccessKey   string        `mapstructure:"R2_SECRET_ACCESS_KEY"`
	R2BucketName        string        `mapstructure:"R2_BUCKET_NAME"`
	R2PublicURL         string        `mapstructure:"R2_PUBLIC_URL"`
	CollegeCode         string        `mapstructure:"COLLEGE_CODE"`
}

func LoadConfig(path string) (Config, error) {
	// Try loading .env first
	viper.SetConfigFile(path + "/.env")
	viper.AutomaticEnv()

	var config Config
	err := viper.ReadInConfig()
	if err != nil {
		// Fallback to app.env
		viper.SetConfigFile(path + "/app.env")
		err = viper.ReadInConfig()
		if err != nil {
			return config, err
		}
	}
	err = viper.Unmarshal(&config)

	return config, err
}
