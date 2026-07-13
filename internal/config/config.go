package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	Port             int
	APIKey           string
	RateLimit        int
	RateWindow       int
	ReadTimeout      int
	WriteTimeout     int
	MaxBodySize      string
	MaxAddressLength int
	AddressDBPath    string
	LocationDBPath   string
}

func Load() *Config {
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	viper.SetDefault("PORT", 7300)
	viper.SetDefault("API_KEY", "")
	viper.SetDefault("RATE_LIMIT", 100)
	viper.SetDefault("RATE_WINDOW", 60)
	viper.SetDefault("READ_TIMEOUT", 5)
	viper.SetDefault("WRITE_TIMEOUT", 10)
	viper.SetDefault("MAX_BODY_SIZE", "1M")
	viper.SetDefault("MAX_ADDRESS_LENGTH", 1000)
	viper.SetDefault("ADDRESS_DB_PATH", "db/address.db")
	viper.SetDefault("LOCATION_DB_PATH", "db/location.db")

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("warning: .env not found, using defaults: %v", err)
	}

	cfg := &Config{
		Port:             viper.GetInt("PORT"),
		APIKey:           viper.GetString("API_KEY"),
		RateLimit:        viper.GetInt("RATE_LIMIT"),
		RateWindow:       viper.GetInt("RATE_WINDOW"),
		ReadTimeout:      viper.GetInt("READ_TIMEOUT"),
		WriteTimeout:     viper.GetInt("WRITE_TIMEOUT"),
		MaxBodySize:      viper.GetString("MAX_BODY_SIZE"),
		MaxAddressLength: viper.GetInt("MAX_ADDRESS_LENGTH"),
		AddressDBPath:    viper.GetString("ADDRESS_DB_PATH"),
		LocationDBPath:   viper.GetString("LOCATION_DB_PATH"),
	}

	log.Printf("config loaded: port=%d, rate_limit=%d, rate_window=%ds", cfg.Port, cfg.RateLimit, cfg.RateWindow)
	return cfg
}
