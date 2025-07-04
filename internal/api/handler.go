package api

import (
	"log/slog"

	"github.com/Unfield/Odin-DNS/internal/config"
	"github.com/Unfield/Odin-DNS/internal/datastore"
	"github.com/Unfield/Odin-DNS/internal/metrics"
)

type Handler struct {
	store  datastore.Driver
	config *config.Config
	logger *slog.Logger
}

func NewHandler(store datastore.Driver, config *config.Config) *Handler {
	return &Handler{
		store:  store,
		config: config,
		logger: slog.Default().WithGroup("API-Handler"),
	}
}

type MetricsHandler struct {
	store              datastore.Driver
	config             *config.Config
	logger             *slog.Logger
	metricsQueryDriver metrics.MetricsQueryDriver
}

func NewMetricsHandler(
	config *config.Config,
	logger *slog.Logger,
	metricsQueryDriver metrics.MetricsQueryDriver,
) *MetricsHandler {
	return &MetricsHandler{
		config:             config,
		logger:             logger,
		metricsQueryDriver: metricsQueryDriver,
	}
}
