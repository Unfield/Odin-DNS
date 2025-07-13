package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	DNS_PORT    int    `json:"dns_port" yaml:"dns_port" xml:"dns_port"`
	DNS_HOST    string `json:"dns_host" yaml:"dns_host" xml:"dns_host"`
	BUFFER_SIZE int    `json:"buffer_size" yaml:"buffer_size" xml:"buffer_size"`

	API_ENABLED bool   `json:"api_enabled" yaml:"api_enabled" xml:"api_enabled"`
	API_PORT    int    `json:"api_port" yaml:"api_port" xml:"api_port"`
	API_HOST    string `json:"api_host" yaml:"api_host" xml:"api_host"`

	MySQL_DSN string `json:"mysql_dsn" yaml:"mysql_dsn" xml:"mysql_dsn"`

	REDIS_HOST     string `json:"redis_host" yaml:"redis_host" xml:"redis_host"`
	REDIS_USERNAME string `json:"redis_username" yaml:"redis_username" xml:"redis_username"`
	REDIS_PASSWORD string `json:"redis_password" yaml:"redis_password" xml:"redis_password"`
	REDIS_DATABASE int    `json:"redis_database" yaml:"redis_database" xml:"redis_database"`

	CORS_ORIGINS []string `json:"cors_origins" yaml:"cors_origins" xml:"cors_origins"`

	CLICKHOUSE_HOST               string        `json:"clickhouse_host" yaml:"clickhouse_host" xml:"clickhouse_host"`
	CLICKHOUSE_DATABASE           string        `json:"clickhouse_database" yaml:"clickhouse_database" xml:"clickhouse_database"`
	CLICKHOUSE_USERNAME           string        `json:"clickhouse_username" yaml:"clickhouse_username" xml:"clickhouse_username"`
	CLICKHOUSE_PASSWORD           string        `json:"clickhouse_password" yaml:"clickhouse_password" xml:"clickhouse_password"`
	CLICKHOUSE_MAX_EXECUTION_TIME int           `json:"clickhouse_max_execution_time" yaml:"clickhouse_max_execution_time" xml:"clickhouse_max_execution_time"`
	CLICKHOUSE_TIMEOUT            int           `json:"clickhouse_timeout" yaml:"clickhouse_timeout" xml:"clickhouse_timeout"`
	CLICKHOUSE_MAX_BATCH_SIZE     int           `json:"clickhouse_max_batch_size" yaml:"clickhouse_max_batch_size" xml:"clickhouse_max_batch_size"`
	CLICKHOUSE_BATCH_INTERVAL     time.Duration `json:"clickhouse_batch_interval" yaml:"clickhouse_batch_interval" xml:"clickhouse_batch_interval"`
}

func DefaultConfig() *Config {
	return &Config{
		DNS_PORT:                      53,
		DNS_HOST:                      "127.0.0.1",
		BUFFER_SIZE:                   512,
		API_ENABLED:                   true,
		API_PORT:                      8080,
		API_HOST:                      "127.0.0.1",
		MySQL_DSN:                     "",
		REDIS_HOST:                    "localhost:6379",
		REDIS_USERNAME:                "default",
		REDIS_PASSWORD:                "",
		REDIS_DATABASE:                0,
		CLICKHOUSE_HOST:               "localhost:9000",
		CLICKHOUSE_DATABASE:           "odindns",
		CLICKHOUSE_USERNAME:           "default",
		CLICKHOUSE_PASSWORD:           "",
		CLICKHOUSE_MAX_EXECUTION_TIME: 60,
		CLICKHOUSE_TIMEOUT:            30,
		CLICKHOUSE_MAX_BATCH_SIZE:     1000,
		CLICKHOUSE_BATCH_INTERVAL:     5,
		CORS_ORIGINS:                  []string{},
	}
}

func LoadConfig() (*Config, error) {
	cfg := DefaultConfig()

	formatCorsString := func(input string) []string {
		if input == "" {
			return []string{}
		}

		var result []string
		for _, part := range strings.Split(input, ",") {
			if trimmed := strings.TrimSpace(part); trimmed != "" {
				result = append(result, trimmed)
			}
		}
		return result
	}

	getCorsArray := func(envVar string, defaultValue []string) []string {
		if value := os.Getenv(envVar); value != "" {
			return formatCorsString(value)
		}
		return defaultValue
	}

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

	getDuration := func(envVar string, defaultValue time.Duration) (time.Duration, error) {
		if valueStr := os.Getenv(envVar); valueStr != "" {
			value, err := strconv.Atoi(valueStr)
			if err != nil {
				return 0, fmt.Errorf("invalid value for environment variable %s: %w", envVar, err)
			}
			return time.Duration(value) * time.Second, nil
		}
		return defaultValue, nil
	}

	var err error

	cfg.DNS_PORT, err = getInt("ODIN_DNS_PORT", cfg.DNS_PORT)
	cfg.DNS_HOST = getString("ODIN_DNS_HOST", cfg.DNS_HOST)
	cfg.BUFFER_SIZE, err = getInt("ODIN_BUFFER_SIZE", cfg.BUFFER_SIZE)

	cfg.API_ENABLED, err = getBool("ODIN_API_ENABLED", cfg.API_ENABLED)
	cfg.API_PORT, err = getInt("ODIN_API_PORT", cfg.API_PORT)
	cfg.API_HOST = getString("ODIN_API_HOST", cfg.API_HOST)

	cfg.CORS_ORIGINS = getCorsArray("ODIN_CORS_ORIGINS", cfg.CORS_ORIGINS)

	cfg.MySQL_DSN = getString("ODIN_MYSQL_DSN", cfg.MySQL_DSN)

	cfg.REDIS_HOST = getString("ODIN_REDIS_HOST", cfg.REDIS_HOST)
	cfg.REDIS_USERNAME = getString("ODIN_REDIS_USERNAME", cfg.REDIS_USERNAME)
	cfg.REDIS_PASSWORD = getString("ODIN_REDIS_PASSWORD", cfg.REDIS_PASSWORD)
	cfg.REDIS_DATABASE, err = getInt("ODIN_REDIS_DATABASE", cfg.REDIS_DATABASE)

	cfg.CLICKHOUSE_HOST = getString("ODIN_CLICKHOUSE_HOST", cfg.CLICKHOUSE_HOST)
	cfg.CLICKHOUSE_DATABASE = getString("ODIN_CLICKHOUSE_DATABASE", cfg.CLICKHOUSE_DATABASE)
	cfg.CLICKHOUSE_USERNAME = getString("ODIN_CLICKHOUSE_USERNAME", cfg.CLICKHOUSE_USERNAME)
	cfg.CLICKHOUSE_PASSWORD = getString("ODIN_CLICKHOUSE_PASSWORD", cfg.CLICKHOUSE_PASSWORD)
	cfg.CLICKHOUSE_MAX_EXECUTION_TIME, err = getInt("ODIN_CLICKHOUSE_MAX_EXECUTION_TIME", cfg.CLICKHOUSE_MAX_EXECUTION_TIME)
	cfg.CLICKHOUSE_TIMEOUT, err = getInt("ODIN_CLICKHOUSE_TIMEOUT", cfg.CLICKHOUSE_TIMEOUT)
	cfg.CLICKHOUSE_MAX_BATCH_SIZE, err = getInt("ODIN_CLICKHOUSE_MAX_BATCH_SIZE", cfg.CLICKHOUSE_MAX_BATCH_SIZE)
	cfg.CLICKHOUSE_BATCH_INTERVAL, err = getDuration("ODIN_CLICKHOUSE_BATCH_INTERVAL", cfg.CLICKHOUSE_BATCH_INTERVAL)

	if err != nil {
		return nil, fmt.Errorf("error loading configuration: %w", err)
	}

	return cfg, nil
}
