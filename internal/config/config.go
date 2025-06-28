package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	DNS_PORT    int `json:"dns_port" yaml:"dns_port" xml:"dns_port"`
	BUFFER_SIZE int `json:"buffer_size" yaml:"buffer_size" xml:"buffer_size"`

	API_ENABLED bool   `json:"api_enabled" yaml:"api_enabled" xml:"api_enabled"`
	API_PORT    int    `json:"api_port" yaml:"api_port" xml:"api_port"`
	API_HOST    string `json:"api_host" yaml:"api_host" xml:"api_host"`

	MySQL_DSN string `json:"mysql_dsn" yaml:"mysql_dsn" xml:"mysql_dsn"`

	REDIS_HOST     string `json:"redis_host" yaml:"redis_host" xml:"redis_host"`
	REDIS_USERNAME string `json:"redis_username" yaml:"redis_username" xml:"redis_username"`
	REDIS_PASSWORD string `json:"redis_password" yaml:"redis_password" xml:"redis_password"`
	REDIS_DATABASE int    `json:"redis_database" yaml:"redis_database" xml:"redis_database"`
}

func DefaultConfig() *Config {
	return &Config{
		DNS_PORT:       53,
		BUFFER_SIZE:    512,
		API_ENABLED:    true,
		API_PORT:       8080,
		API_HOST:       "127.0.0.1",
		MySQL_DSN:      "admin:admin@tcp(127.0.0.1:3306)/odindns?parseTime=true",
		REDIS_HOST:     "localhost:6379",
		REDIS_USERNAME: "default",
		REDIS_PASSWORD: "",
		REDIS_DATABASE: 0,
	}
}

func LoadConfig() (*Config, error) {
	cfg := DefaultConfig()

	getString := func(envVar string, defaultValue string) string {
		if value := os.Getenv(envVar); value != "" {
			return value
		}
		return defaultValue
	}

	getInt := func(envVar string, defaultValue int) (int, error) {
		if valueStr := os.Getenv(envVar); valueStr != "" {
			value, err := strconv.Atoi(valueStr)
			if err != nil {
				return 0, fmt.Errorf("invalid value for environment variable %s: %w", envVar, err)
			}
			return value, nil
		}
		return defaultValue, nil
	}

	getBool := func(envVar string, defaultValue bool) (bool, error) {
		if valueStr := os.Getenv(envVar); valueStr != "" {
			value, err := strconv.ParseBool(valueStr)
			if err != nil {
				return false, fmt.Errorf("invalid value for environment variable %s: %w", envVar, err)
			}
			return value, nil
		}
		return defaultValue, nil
	}

	var err error

	cfg.DNS_PORT, err = getInt("ODIN_DNS_PORT", cfg.DNS_PORT)
	if err != nil {
		return nil, err
	}
	cfg.BUFFER_SIZE, err = getInt("ODIN_BUFFER_SIZE", cfg.BUFFER_SIZE)
	if err != nil {
		return nil, err
	}

	cfg.API_ENABLED, err = getBool("ODIN_API_ENABLED", cfg.API_ENABLED)
	if err != nil {
		return nil, err
	}
	cfg.API_PORT, err = getInt("ODIN_API_PORT", cfg.API_PORT)
	if err != nil {
		return nil, err
	}
	cfg.API_HOST = getString("ODIN_API_HOST", cfg.API_HOST)

	cfg.MySQL_DSN = getString("ODIN_MYSQL_DSN", cfg.MySQL_DSN)

	cfg.REDIS_HOST = getString("ODIN_REDIS_HOST", cfg.REDIS_HOST)
	cfg.REDIS_USERNAME = getString("ODIN_REDIS_USERNAME", cfg.REDIS_USERNAME)
	cfg.REDIS_PASSWORD = getString("ODIN_REDIS_PASSWORD", cfg.REDIS_PASSWORD)
	cfg.REDIS_DATABASE, err = getInt("ODIN_REDIS_DATABASE", cfg.REDIS_DATABASE)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
