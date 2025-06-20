package main

import (
	"log/slog"

	"github.com/Unfield/Odin-DNS/internal/config"
	"github.com/Unfield/Odin-DNS/internal/server"
)

func main() {
	config, err := config.LoadConfig()
	if err != nil {
		slog.Error("Error loading configuration", "error", err)
		return
	}

	server.StartServer(config)
}
