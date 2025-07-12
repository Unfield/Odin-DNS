package api

import (
	"net/http"
	"strconv"

	"github.com/Unfield/Odin-DNS/internal/models"
	"github.com/Unfield/Odin-DNS/internal/util"
)

// GetMonthlyRequestsErrorsHandler retrieves monthly requests and errors data
// @Summary Get Monthly Requests and Errors
// @Description Returns time series data for monthly DNS requests and errors
// @Tags metrics
// @Security BearerAuth
// @Produce json
// @Success 200 {array} models.TimeSeriesData "Monthly requests and errors data retrieved successfully"
// @Failure 500 {object} models.GenericErrorResponse "Failed to retrieve monthly requests and errors data"
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

// GetDailyRequestsErrorsHandler retrieves daily requests and errors data
// @Summary Get Daily Requests and Errors
// @Description Returns time series data for daily DNS requests and errors
// @Tags metrics
// @Security BearerAuth
// @Produce json
// @Success 200 {array} models.TimeSeriesData "Daily requests and errors data retrieved successfully"
// @Failure 500 {object} models.GenericErrorResponse "Failed to retrieve daily requests and errors data"
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

// GetOverallSummaryMetricsHandler retrieves overall summary metrics
// @Summary Get Overall Summary Metrics
// @Description Returns overall summary metrics including response times, cache hit rates, and request counts
// @Tags metrics
// @Security BearerAuth
// @Produce json
// @Param hours query int false "Number of hours to look back (default: 24)" default(24)
// @Success 200 {object} models.GlobalAvgMetrics "Overall summary metrics retrieved successfully"
// @Failure 500 {object} models.GenericErrorResponse "Failed to retrieve overall summary metrics"
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

// GetTopDomainsHandler retrieves top queried domains
// @Summary Get Top Domains
// @Description Returns the most frequently queried domains
// @Tags metrics
// @Security BearerAuth
// @Produce json
// @Param limit query int false "Number of top domains to return (default: 10)" default(10)
// @Success 200 {array} models.TopNData "Top domains retrieved successfully"
// @Failure 500 {object} models.GenericErrorResponse "Failed to retrieve top domains"
// @Router /api/v1/metrics/top-domains [get]
func (h *MetricsHandler) GetTopDomainsHandler(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 10
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

// GetRcodeDistributionHandler retrieves RCODE distribution data
// @Summary Get RCODE Distribution
// @Description Returns the distribution of DNS response codes
// @Tags metrics
// @Security BearerAuth
// @Produce json
// @Success 200 {array} models.RcodeData "RCODE distribution data retrieved successfully"
// @Failure 500 {object} models.GenericErrorResponse "Failed to retrieve RCODE distribution data"
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

// GetQPMHandler retrieves queries per minute data
// @Summary Get Queries Per Minute
// @Description Returns time series data for queries per minute
// @Tags metrics
// @Security BearerAuth
// @Produce json
// @Success 200 {array} models.TimeSeriesData "Queries per minute data retrieved successfully"
// @Failure 500 {object} models.GenericErrorResponse "Failed to retrieve queries per minute data"
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
