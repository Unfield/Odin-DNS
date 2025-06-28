package api

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/Unfield/Odin-DNS/internal/api/middleware"
	"github.com/Unfield/Odin-DNS/internal/config"
	mysql "github.com/Unfield/Odin-DNS/internal/datastore/MySQL"
	"github.com/Unfield/Odin-DNS/internal/util"
)

func StartRouter(config *config.Config) {
	logger := slog.Default().WithGroup("API")

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", HealthCheckHandler)

	mux.Handle("GET /swagger/", middleware.SwaggerHandler())
	mux.HandleFunc("GET /swagger", middleware.SwaggerRedirect)
	logger.Info("Swagger UI enabled", "url", fmt.Sprintf("http://%s:%d/swagger/", config.API_HOST, config.API_PORT))

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

	corsMux := middleware.CORS(mux)

	logger.Info("Odin DNS API running", "port", config.API_PORT)
	http.ListenAndServe(fmt.Sprintf("%s:%d", config.API_HOST, config.API_PORT), corsMux)
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

type HealthResponse struct {
	Status    string `json:"status" example:"OK"`
	Message   string `json:"message,omitempty" example:"API is healthy and operational"`
	Timestamp string `json:"timestamp" example:"2025-06-28T12:00:00Z"`
}

// HealthCheckHandler handles the health check endpoint
// @Summary API Health Check
// @Description Returns the operational status of the API.
// @Tags health
// @Produce json
// @Success 200 {object} HealthResponse "API is healthy"
// @Router /health [get]
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status:    "OK",
		Message:   "API is healthy and operational",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode health response", http.StatusInternalServerError)
	}
}
