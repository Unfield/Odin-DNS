package main

import (
	"log/slog"

	"github.com/Unfield/Odin-DNS/internal/api"
	"github.com/Unfield/Odin-DNS/internal/config"
	"github.com/Unfield/Odin-DNS/internal/server"
	_ "github.com/joho/godotenv/autoload"

	_ "github.com/Unfield/Odin-DNS/docs"
)

//	@title			Odin DNS API
//	@version		1.0
//	@description	Advanced DNS server with REST API for managing DNS records and zones.
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	Odin DNS Support
//	@contact.url	https://github.com/Unfield/Odin-DNS
//	@contact.email	support@odindns.local

//	@license.name	MIT License
//	@license.url	https://github.com/Unfield/Odin-DNS/blob/main/LICENSE

// @host		localhost:8080
// @BasePath	/
// @schemes	http https

// @securityDefinitions.apikey DemoKey
// @in query
// @name demo_key
// @description Demo API key required for all endpoints

func main() {
	config, err := config.LoadConfig()
	if err != nil {
		slog.Error("Error loading configuration", "error", err)
		return
	}

	if config.DEMO_KEY == "changeme" {
		slog.Error("Demo key is not set. Please set the DEMO_KEY in the environment variables or config file.")
		return
	}

	if config.API_ENABLED {
		go api.StartRouter(config)
	}

	server.StartServer(config)
}
