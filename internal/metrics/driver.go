package metrics

import (
	"time"

	"github.com/Unfield/Odin-DNS/internal/models"
)

type MetricsIngestionDriver interface {
	Collect(DNSMetric)
	Close() error
}

type MetricsQueryDriver interface {
	GetMonthlyRequestsErrors() ([]models.TimeSeriesData, error)
	GetDailyRequestsErrors() ([]models.TimeSeriesData, error)
	GetOverallSummaryMetrics(hours int) (*models.GlobalAvgMetrics, error)
	GetTopDomains(limit int) ([]models.TopNData, error)
	GetRcodeDistribution() ([]models.RcodeData, error)
	GetQPM() ([]models.TimeSeriesData, error)
	Close() error
}

type DNSMetric struct {
	Timestamp      time.Time
	IP             string
	Domain         string
	QueryType      string
	Success        uint8
	ErrorMessage   string
	ResponseTimeMs float64
	CacheHit       uint8
	Rcode          uint8
}
