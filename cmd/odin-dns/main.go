package main

import (
	"log/slog"

	"github.com/Unfield/Odin-DNS/internal/api"
	"github.com/Unfield/Odin-DNS/internal/config"
	"github.com/Unfield/Odin-DNS/internal/server"
	_ "github.com/joho/godotenv/autoload"

	_ "github.com/Unfield/Odin-DNS/docs"
)

// Package api provides the REST API for Odin DNS management system
// @title Odin DNS API
// @version 1.0
// @description REST API for managing DNS zones and records, with comprehensive metrics and monitoring
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.email support@odin-dns.com
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
// @host api.odin-demo.drinkuth.online
// @BasePath /
// @schemes http https
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter the token with the `Bearer: ` prefix, e.g. "Bearer abcde12345".
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
