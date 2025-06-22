package main

import (
	"log/slog"

	"github.com/Unfield/Odin-DNS/internal/api"
	"github.com/Unfield/Odin-DNS/internal/config"
	"github.com/Unfield/Odin-DNS/internal/server"
)

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
