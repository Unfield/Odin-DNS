package api

import (
	"net/http"
	"strconv"

	"github.com/Unfield/Odin-DNS/internal/models" // Assuming your models are here
	// Assuming metrics.MetricsQueryDriver and types are here
	"github.com/Unfield/Odin-DNS/internal/util"
)

// GetMonthlyRequestsErrorsHandler retrieves monthly DNS requests and errors data.
// @Summary Get Monthly DNS Metrics
// @Description Retrieves aggregated data for total DNS requests and errors per month.
// @Tags /api/v1/metrics
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.TimeSeriesData "Monthly DNS requests and errors data"
// @Failure 401 {object} models.GenericErrorResponse "Unauthorized - invalid or missing session token"
// @Failure 500 {object} models.GenericErrorResponse "Internal server error"
// @Router /api/v1/metrics/requests/errors/monthly [get]
func (h *MetricsHandler) GetMonthlyRequestsErrorsHandler(w http.ResponseWriter, r *http.Request) {
	data, err := h.metricsQueryDriver.GetMonthlyRequestsErrors()
	if err != nil {
		util.RespondWithJSON(w, http.StatusInternalServerError, models.GenericErrorResponse{
			Error:        true,
			ErrorMessage: "Failed to retrieve monthly requests and errors data",
		})
		return
	}
	util.RespondWithJSON(w, http.StatusOK, data)
}

// GetDailyRequestsErrorsHandler retrieves daily DNS requests and errors data.
// @Summary Get Daily DNS Metrics
// @Description Retrieves aggregated data for total DNS requests and errors per day.
// @Tags /api/v1/metrics
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.TimeSeriesData "Daily DNS requests and errors data"
// @Failure 401 {object} models.GenericErrorResponse "Unauthorized - invalid or missing session token"
// @Failure 500 {object} models.GenericErrorResponse "Internal server error"
// @Router /api/v1/metrics/requests/errors/daily [get]
func (h *MetricsHandler) GetDailyRequestsErrorsHandler(w http.ResponseWriter, r *http.Request) {
	data, err := h.metricsQueryDriver.GetDailyRequestsErrors()
	if err != nil {
		util.RespondWithJSON(w, http.StatusInternalServerError, models.GenericErrorResponse{
			Error:        true,
			ErrorMessage: "Failed to retrieve daily requests and errors data",
		})
		return
	}
	util.RespondWithJSON(w, http.StatusOK, data)
}

// GetOverallSummaryMetricsHandler retrieves overall summary metrics for DNS operations.
// @Summary Get Overall DNS Summary Metrics
// @Description Retrieves a single object containing aggregated summary metrics like average response times, cache hit percentage, total requests, and total errors for a specified time period.
// @Tags /api/v1/metrics
// @Produce json
// @Security BearerAuth
// @Param hours query int false "Number of hours to look back for metrics (default: 24)" default(24) minimum(1) maximum(8760)
// @Success 200 {object} models.GlobalAvgMetrics "Overall DNS summary metrics"
// @Failure 401 {object} models.GenericErrorResponse "Unauthorized - invalid or missing session token"
// @Failure 500 {object} models.GenericErrorResponse "Internal server error"
// @Router /api/v1/metrics/summary [get]
func (h *MetricsHandler) GetOverallSummaryMetricsHandler(w http.ResponseWriter, r *http.Request) {
	hours, err := strconv.Atoi(r.URL.Query().Get("hours"))
	if err != nil || hours <= 0 {
		hours = 24
	}
	data, err := h.metricsQueryDriver.GetOverallSummaryMetrics(hours)
	if err != nil {
		util.RespondWithJSON(w, http.StatusInternalServerError, models.GenericErrorResponse{
			Error:        true,
			ErrorMessage: "Failed to retrieve overall summary metrics",
		})
		return
	}
	util.RespondWithJSON(w, http.StatusOK, data)
}

// GetTopDomainsHandler retrieves the top N most queried domains.
// @Summary Get Top Queried Domains
// @Description Retrieves a list of the top N most frequently queried domains, sorted by count in descending order. Default limit is 10.
// @Tags /api/v1/metrics
// @Produce json
// @Param limit query int false "Number of top domains to retrieve" default(10) example(5)
// @Security BearerAuth
// @Success 200 {array} models.TopNData "List of top domains and their query counts"
// @Failure 401 {object} models.GenericErrorResponse "Unauthorized - invalid or missing session token"
// @Failure 500 {object} models.GenericErrorResponse "Internal server error"
// @Router /api/v1/metrics/top-domains [get]
func (h *MetricsHandler) GetTopDomainsHandler(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 10 // Default limit
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	data, err := h.metricsQueryDriver.GetTopDomains(limit)
	if err != nil {
		util.RespondWithJSON(w, http.StatusInternalServerError, models.GenericErrorResponse{
			Error:        true,
			ErrorMessage: "Failed to retrieve top domains",
		})
		return
	}
	util.RespondWithJSON(w, http.StatusOK, data)
}

// GetRcodeDistributionHandler retrieves the distribution of DNS response codes.
// @Summary Get RCODE Distribution
// @Description Retrieves the count of queries for each DNS response code (e.g., NOERROR, NXDOMAIN, SERVFAIL).
// @Tags /api/v1/metrics
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.RcodeData "List of RCODEs and their counts"
// @Failure 401 {object} models.GenericErrorResponse "Unauthorized - invalid or missing session token"
// @Failure 500 {object} models.GenericErrorResponse "Internal server error"
// @Router /api/v1/metrics/rcode-distribution [get]
func (h *MetricsHandler) GetRcodeDistributionHandler(w http.ResponseWriter, r *http.Request) {
	data, err := h.metricsQueryDriver.GetRcodeDistribution()
	if err != nil {
		util.RespondWithJSON(w, http.StatusInternalServerError, models.GenericErrorResponse{
			Error:        true,
			ErrorMessage: "Failed to retrieve RCODE distribution data",
		})
		return
	}
	util.RespondWithJSON(w, http.StatusOK, data)
}

// GetQPMHandler retrieves queries per minute (QPM) data.
// @Summary Get Queries Per Minute (QPM)
// @Description Retrieves aggregated data for the number of DNS queries per minute over a recent period.
// @Tags /api/v1/metrics
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.TimeSeriesData "Queries per minute data"
// @Failure 401 {object} models.GenericErrorResponse "Unauthorized - invalid or missing session token"
// @Failure 500 {object} models.GenericErrorResponse "Internal server error"
// @Router /api/v1/metrics/qpm [get]
func (h *MetricsHandler) GetQPMHandler(w http.ResponseWriter, r *http.Request) {
	data, err := h.metricsQueryDriver.GetQPM()
	if err != nil {
		util.RespondWithJSON(w, http.StatusInternalServerError, models.GenericErrorResponse{
			Error:        true,
			ErrorMessage: "Failed to retrieve queries per minute data",
		})
		return
	}
	util.RespondWithJSON(w, http.StatusOK, data)
}
