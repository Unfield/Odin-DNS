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
	redis "github.com/Unfield/Odin-DNS/internal/datastore/Redis"
)

func StartRouter(config *config.Config) {
	logger := slog.Default().WithGroup("API")

	mux := http.NewServeMux()

	mysqlDriver, err := mysql.NewMySQLDriver(config.MySQL_DSN)
	if err != nil {
		logger.Error("Failed to connect to MySQL", "error", err)
		return
	}

	cacheDriver := redis.NewRedisCacheDriver(mysqlDriver, config.REDIS_HOST, config.REDIS_USERNAME, config.REDIS_PASSWORD, config.REDIS_DATABASE)

	corsConfig := middleware.CORSConfig{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization", "X-Requested-With", "Accept"},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           3600,
	}

	chain := middleware.New(
		middleware.CORS(corsConfig),
		middleware.RequestID(),
		middleware.Logger(),
		middleware.Recovery(),
		middleware.Timeout(30*time.Second),
	)

	protectedChain := chain.Use(
		middleware.AuthMiddleware(cacheDriver),
	)

	rateLimiter := middleware.NewRateLimiter(10, time.Minute)
	apiChain := chain.Use(rateLimiter.Middleware())

	mux.HandleFunc("GET /health", chain.ThenFunc(HealthCheckHandler).ServeHTTP)

	// Swagger UI
	mux.Handle("GET /swagger/", chain.Then(middleware.SwaggerHandler()))
	mux.HandleFunc("GET /swagger", chain.ThenFunc(middleware.SwaggerRedirect).ServeHTTP)
	logger.Info("Swagger UI enabled", "url", fmt.Sprintf("http://%s:%d/swagger/", config.API_HOST, config.API_PORT))

	handler := NewHandler(mysqlDriver, config)

	mux.Handle("POST /api/v1/login", apiChain.ThenFunc(http.HandlerFunc(handler.LoginHandler)))
	mux.Handle("POST /api/v1/register", apiChain.ThenFunc(http.HandlerFunc(handler.RegisterHandler)))
	mux.Handle("POST /api/v1/logout", protectedChain.ThenFunc(http.HandlerFunc(handler.LogoutHandler)))
	mux.Handle("GET /api/v1/user/{session_id}", protectedChain.ThenFunc(http.HandlerFunc(handler.GetUserHandler)))

	/*
		// Zone management routes
		mux.Handle("POST /api/v1/zone", protectedChain.ThenFunc(http.HandlerFunc(handler.CreateZoneHandler)))
		mux.Handle("GET /api/v1/zone/records/{session_id}", protectedChain.ThenFunc(http.HandlerFunc(handler.GetZoneRecordsHandler)))
		mux.Handle("DELETE /api/v1/zone", protectedChain.ThenFunc(http.HandlerFunc(handler.DeleteZoneHandler)))
		// Record management routes
		mux.Handle("POST /api/v1/record", protectedChain.ThenFunc(http.HandlerFunc(handler.CreateRecordHandler)))
	*/

	logger.Info("Odin DNS API running", "port", config.API_PORT)
	http.ListenAndServe(fmt.Sprintf("%s:%d", config.API_HOST, config.API_PORT), mux)
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
