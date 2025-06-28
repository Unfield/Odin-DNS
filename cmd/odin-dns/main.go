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
//	@description	Advanced DNS server with REST API for managing DNS records and zones. This API provides comprehensive DNS management capabilities including user authentication, zone management, and DNS record operations.
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	Odin DNS Support
//	@contact.url	https://github.com/Unfield/Odin-DNS
//	@contact.email	support@odindns.local

//	@license.name	MIT License
//	@license.url	https://github.com/Unfield/Odin-DNS/blob/main/LICENSE

// @host		localhost:8080
// @BasePath	/
// @schemes	http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter your bearer token in the format: Bearer {token}

// @tag.name health
// @tag.description Health check endpoints

// @tag.name authentication
// @tag.description User authentication and session management

// @tag.name user
// @tag.description User profile and account management

// @tag.name zones
// @tag.description DNS zone management operations

// @tag.name records
// @tag.description DNS record management operations

func main() {
	config, err := config.LoadConfig()
	if err != nil {
		slog.Error("Error loading configuration", "error", err)
		return
	}

	if config.API_ENABLED {
		go api.StartRouter(config)
	}

	server.StartServer(config)
}
