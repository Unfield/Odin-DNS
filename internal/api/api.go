package api

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/Unfield/Odin-DNS/internal/config"
	mysql "github.com/Unfield/Odin-DNS/internal/datastore/MySQL"
	"github.com/Unfield/Odin-DNS/internal/util"
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
	mux.Handle("POST /api/v1/login", DemoKeyChecker(config, logger, http.HandlerFunc(handler.LoginHandler)))
	mux.Handle("POST /api/v1/register", DemoKeyChecker(config, logger, http.HandlerFunc(handler.RegisterHandler)))
	mux.Handle("POST /api/v1/logout", DemoKeyChecker(config, logger, http.HandlerFunc(handler.LogoutHandler)))
	mux.Handle("GET /api/v1/user/{session_id}", DemoKeyChecker(config, logger, http.HandlerFunc(handler.GetUserHandler)))

	// Zone management routes
	mux.Handle("POST /api/v1/zone", DemoKeyChecker(config, logger, http.HandlerFunc(handler.CreateZoneHandler)))
	mux.Handle("GET /api/v1/zone/records/{session_id}", DemoKeyChecker(config, logger, http.HandlerFunc(handler.GetZoneRecordsHandler)))
	mux.Handle("DELETE /api/v1/zone", DemoKeyChecker(config, logger, http.HandlerFunc(handler.DeleteZoneHandler)))
	// Record management routes
	mux.Handle("POST /api/v1/record", DemoKeyChecker(config, logger, http.HandlerFunc(handler.CreateRecordHandler)))

	logger.Info("Odin DNS API running", "port", config.API_PORT)
	http.ListenAndServe(fmt.Sprintf("%s:%d", config.API_HOST, config.API_PORT), mux)
}

func DemoKeyChecker(config *config.Config, logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		checkSuccessfull := util.CheckForDemoKey(r.URL.Query(), w, config.DEMO_KEY)
		if !checkSuccessfull {
			logger.Info("Get user attempt with invalid demo key")
			return
		}
		next.ServeHTTP(w, r)
	})
}
