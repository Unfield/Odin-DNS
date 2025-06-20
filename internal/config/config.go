package config

type Config struct {
	DNS_PORT    int `json:"dns_port" yaml:"dns_port" xml:"dns_port"`
	BUFFER_SIZE int `json:"buffer_size" yaml:"buffer_size" xml:"buffer_size"`
}

func DefaultConfig() (*Config, error) {
	return &Config{
		DNS_PORT:    53,
		BUFFER_SIZE: 512,
	}, nil
}

func LoadConfig() (*Config, error) {
	return DefaultConfig()
}

func (c *Config) SetDNSPort(port int) {
	c.DNS_PORT = port
}

func (c *Config) SetBufferSize(size int) {
	c.BUFFER_SIZE = size
}

func (c *Config) GetDNSPort() int {
	return c.DNS_PORT
}

func (c *Config) GetBufferSize() int {
	return c.BUFFER_SIZE
}

func (c *Config) IsValid() bool {
	return c.DNS_PORT > 0 && c.BUFFER_SIZE > 0
}
