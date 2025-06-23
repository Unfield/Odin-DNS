package config

type Config struct {
	DNS_PORT    int `json:"dns_port" yaml:"dns_port" xml:"dns_port"`
	BUFFER_SIZE int `json:"buffer_size" yaml:"buffer_size" xml:"buffer_size"`

	API_ENABLED bool   `json:"api_enabled" yaml:"api_enabled" xml:"api_enabled"`
	API_PORT    int    `json:"api_port" yaml:"api_port" xml:"api_port"`
	API_HOST    string `json:"api_host" yaml:"api_host" xml:"api_host"`

	DEMO_KEY string `json:"demo_key" yaml:"demo_key" xml:"demo_key"`

	MySQL_DSN string `json:"mysql_dsn" yaml:"mysql_dsn" xml:"mysql_dsn"`

	REDIS_HOST     string `json:"redis_host" yaml:"redis_host" xml:"redis_host"`
	REDIS_USERNAME string `json:"redis_username" yaml:"redis_username" xml:"redis_username"`
	REDIS_PASSWORD string `json:"redis_password" yaml:"redis_password" xml:"redis_password"`
	REDIS_DATABASE int    `json:"redis_database" yaml:"redis_database" xml:"redis_database"`
}

func DefaultConfig() (*Config, error) {
	return &Config{
		DNS_PORT:       53,
		BUFFER_SIZE:    512,
		API_ENABLED:    true,
		API_PORT:       8080,
		API_HOST:       "127.0.0.1",
		DEMO_KEY:       "8icGOXsNqQrIT0d6Nbhk6Bb9oSfkztvq",
		MySQL_DSN:      "admin:admin@tcp(127.0.0.1:3306)/odindns?parseTime=true",
		REDIS_HOST:     "localhost:6379",
		REDIS_USERNAME: "default",
		REDIS_PASSWORD: "",
		REDIS_DATABASE: 0,
	}, nil
}

func LoadConfig() (*Config, error) {
	return DefaultConfig()
}
