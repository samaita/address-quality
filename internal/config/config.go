package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	Port         int
	APIKey       string
	RateLimit    int
	RateWindow   int
	ReadTimeout  int
	WriteTimeout int
}

var Cfg *Config

func Init() {
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	viper.SetDefault("PORT", 7300)
	viper.SetDefault("API_KEY", "")
	viper.SetDefault("RATE_LIMIT", 100)
	viper.SetDefault("RATE_WINDOW", 60)
	viper.SetDefault("READ_TIMEOUT", 5)
	viper.SetDefault("WRITE_TIMEOUT", 10)

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("warning: .env not found, using defaults: %v", err)
	}

	Cfg = &Config{
		Port:         viper.GetInt("PORT"),
		APIKey:       viper.GetString("API_KEY"),
		RateLimit:    viper.GetInt("RATE_LIMIT"),
		RateWindow:   viper.GetInt("RATE_WINDOW"),
		ReadTimeout:  viper.GetInt("READ_TIMEOUT"),
		WriteTimeout: viper.GetInt("WRITE_TIMEOUT"),
	}

	log.Printf("config loaded: port=%d, rate_limit=%d, rate_window=%ds", Cfg.Port, Cfg.RateLimit, Cfg.RateWindow)
}
