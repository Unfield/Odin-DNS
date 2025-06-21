package api

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/Unfield/Odin-DNS/internal/config"
	mysql "github.com/Unfield/Odin-DNS/internal/datastore/MySQL"
)

func StartRouter(config *config.Config) {
	logger := slog.Default().WithGroup("API")

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	mysqlDriver, err := mysql.NewMySQLDriver(config.MySQL_DSN)
	if err != nil {
		logger.Error("Failed to connect to MySQL", "error", err)
		return
	}

	handler := NewHandler(mysqlDriver, config)

	// User authentication routes
	mux.HandleFunc("POST /api/v1/login", handler.LoginHandler)
	mux.HandleFunc("POST /api/v1/register", handler.RegisterHandler)
	mux.HandleFunc("POST /api/v1/logout", handler.LogoutHandler)
	mux.HandleFunc("GET /api/v1/user", handler.GetUserHandler)

	logger.Info("Odin DNS API running", "port", config.API_PORT)
	http.ListenAndServe(fmt.Sprintf("%s:%d", config.API_HOST, config.API_PORT), mux)
}
