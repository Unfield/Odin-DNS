package metrics

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/Unfield/Odin-DNS/internal/config"
	"github.com/Unfield/Odin-DNS/internal/models"
)

type ClickHouseQueryDriver struct {
	clickHouseDB clickhouse.Conn
	logger       *slog.Logger
}

func NewClickHouseQueryDriver(config *config.Config) MetricsQueryDriver {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{config.CLICKHOUSE_HOST},
		Auth: clickhouse.Auth{
			Database: config.CLICKHOUSE_DATABASE,
			Username: config.CLICKHOUSE_USERNAME,
			Password: config.CLICKHOUSE_PASSWORD,
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		DialTimeout: time.Second * 30,
	})
	if err != nil {
		slog.Error("Failed to connect to ClickHouse", "error", err)
		return nil
	}
	if err := conn.Ping(context.Background()); err != nil {
		slog.Error("Failed to ping ClickHouse after connection", "error", err)
		return nil
	}

	driver := &ClickHouseQueryDriver{
		clickHouseDB: conn,
		logger:       slog.Default().WithGroup("Metrics"),
	}
	return driver
}

func (d *ClickHouseQueryDriver) Close() error {
	if d.clickHouseDB != nil {
		if err := d.clickHouseDB.Close(); err != nil {
			d.logger.Error("Error closing ClickHouse connection", "error", err)
			return err
		}
		d.logger.Info("ClickHouse connection closed successfully.")
	}
	return nil
}

func (d *ClickHouseQueryDriver) GetMonthlyRequestsErrors() ([]models.TimeSeriesData, error) {
	rows, err := d.clickHouseDB.Query(context.Background(), `
		SELECT
			toStartOfMonth(timestamp) as time,
			sum(success) as requests,
			sum(1 - success) as errors
		FROM dns_metrics
		GROUP BY time
		ORDER BY time ASC;
	`)
	if err != nil {
		d.logger.Error("Failed to query monthly requests/errors", "error", err)
		return nil, fmt.Errorf("failed to query monthly requests/errors: %w", err)
	}
	defer rows.Close()

	var results []models.TimeSeriesData
	for rows.Next() {
		var data models.TimeSeriesData
		if err := rows.Scan(&data.Time, &data.Requests, &data.Errors); err != nil {
			d.logger.Error("Failed to scan monthly requests/errors row", "error", err)
			return nil, fmt.Errorf("failed to scan monthly requests/errors row: %w", err)
		}
		results = append(results, data)
	}
	return results, rows.Err()
}

func (d *ClickHouseQueryDriver) GetDailyRequestsErrors() ([]models.TimeSeriesData, error) {
	rows, err := d.clickHouseDB.Query(context.Background(), `
		SELECT
			toStartOfDay(timestamp) as time,
			sum(success) as requests,
			sum(1 - success) as errors
		FROM dns_metrics
		GROUP BY time
		ORDER BY time ASC;
	`)
	if err != nil {
		d.logger.Error("Failed to query daily requests/errors", "error", err)
		return nil, fmt.Errorf("failed to query daily requests/errors: %w", err)
	}
	defer rows.Close()

	var results []models.TimeSeriesData
	for rows.Next() {
		var data models.TimeSeriesData
		if err := rows.Scan(&data.Time, &data.Requests, &data.Errors); err != nil {
			d.logger.Error("Failed to scan daily requests/errors row", "error", err)
			return nil, fmt.Errorf("failed to scan daily requests/errors row: %w", err)
		}
		results = append(results, data)
	}
	return results, rows.Err()
}

func (d *ClickHouseQueryDriver) GetOverallSummaryMetrics(hours int) (*models.GlobalAvgMetrics, error) {
	row := d.clickHouseDB.QueryRow(context.Background(), `
		SELECT
			if(count(*) > 0, avg(response_time_ms), 0) as avg_response_time_ms,
			if(countIf(success = 1) > 0, avgIf(response_time_ms, success = 1), 0) as avg_success_response_time_ms,
			if(countIf(success = 0) > 0, avgIf(response_time_ms, success = 0), 0) as avg_error_response_time_ms,
			if(count(*) > 0, (countIf(cache_hit = 1) * 100.0) / count(*), 0) as cache_hit_percentage,
			count(*) as total_requests,
			countIf(success = 0) as total_errors
		FROM dns_metrics
		WHERE timestamp >= now() - INTERVAL ? HOUR
	`, hours)

	var metrics models.GlobalAvgMetrics
	err := row.Scan(
		&metrics.AvgResponseTimeMs,
		&metrics.AvgSuccessResponseTimeMs,
		&metrics.AvgErrorResponseTimeMs,
		&metrics.CacheHitPercentage,
		&metrics.TotalRequests,
		&metrics.TotalErrors,
	)

	if err != nil {
		d.logger.Error("Failed to scan global avg metrics with time filter", "error", err)
		return nil, fmt.Errorf("failed to scan global avg metrics with time filter: %w", err)
	}

	if math.IsNaN(metrics.AvgResponseTimeMs) {
		metrics.AvgResponseTimeMs = 0
	}
	if math.IsNaN(metrics.AvgSuccessResponseTimeMs) {
		metrics.AvgSuccessResponseTimeMs = 0
	}
	if math.IsNaN(metrics.AvgErrorResponseTimeMs) {
		metrics.AvgErrorResponseTimeMs = 0
	}
	if math.IsNaN(metrics.CacheHitPercentage) {
		metrics.CacheHitPercentage = 0
	}

	return &metrics, nil
}

func (d *ClickHouseQueryDriver) GetTopDomains(limit int) ([]models.TopNData, error) {
	rows, err := d.clickHouseDB.Query(context.Background(), `
		SELECT
			domain,
			count(*) as count
		FROM dns_metrics
		GROUP BY domain
		ORDER BY count DESC
		LIMIT ?
	`, limit)
	if err != nil {
		d.logger.Error("Failed to query top domains", "error", err)
		return nil, fmt.Errorf("failed to query top domains: %w", err)
	}
	defer rows.Close()
	var results []models.TopNData
	for rows.Next() {
		var data models.TopNData
		if err := rows.Scan(&data.Name, &data.Count); err != nil {
			d.logger.Error("Failed to scan top domains row", "error", err)
			return nil, fmt.Errorf("failed to scan top domains row: %w", err)
		}
		results = append(results, data)
	}
	if err := rows.Err(); err != nil {
		d.logger.Error("Error iterating over top domains rows", "error", err)
		return nil, fmt.Errorf("error iterating over top domains rows: %w", err)
	}
	if len(results) == 0 {
		d.logger.Info("No top domains found")
		return nil, nil
	}
	return results, nil
}

func (d *ClickHouseQueryDriver) GetRcodeDistribution() ([]models.RcodeData, error) {
	rows, err := d.clickHouseDB.Query(context.Background(), `
		SELECT
			rcode,
			count(*) as count,
			CASE rcode
				WHEN 0 THEN 'NOERROR'
				WHEN 1 THEN 'FORMERR'
				WHEN 2 THEN 'SERVFAIL'
				WHEN 3 THEN 'NXDOMAIN'
				WHEN 4 THEN 'NOTIMP'
				WHEN 5 THEN 'REFUSED'
				WHEN 6 THEN 'YXDOMAIN'
				WHEN 7 THEN 'YXRRSET'
				WHEN 8 THEN 'NXRRSET'
				WHEN 9 THEN 'NOTAUTH'
				WHEN 10 THEN 'NOTZONE'
				WHEN 11 THEN 'BADVERS'
				WHEN 12 THEN 'BADKEY'
				WHEN 13 THEN 'BADTIME'
				WHEN 14 THEN 'BADMODE'
				WHEN 15 THEN 'BADNAME'
				WHEN 16 THEN 'BADALG'
				WHEN 17 THEN 'BADTRUNC'
				WHEN 18 THEN 'BADCOOKIE'
				ELSE 'UNKNOWN'
			END as rcode_name
		FROM dns_metrics
		GROUP BY rcode
		ORDER BY count DESC;
	`)

	if err != nil {
		d.logger.Error("Failed to query RCODE distribution", "error", err)
		return nil, fmt.Errorf("failed to query RCODE distribution: %w", err)
	}
	defer rows.Close()

	var results []models.RcodeData
	for rows.Next() {
		var data models.RcodeData
		if err := rows.Scan(&data.Rcode, &data.Count, &data.RcodeName); err != nil {
			d.logger.Error("Failed to scan RCODE distribution row", "error", err)
			return nil, fmt.Errorf("failed to scan RCODE distribution row: %w", err)
		}
		results = append(results, data)
	}
	if err := rows.Err(); err != nil {
		d.logger.Error("Error iterating over RCODE distribution rows", "error", err)
		return nil, fmt.Errorf("error iterating over RCODE distribution rows: %w", err)
	}
	if len(results) == 0 {
		d.logger.Info("No RCODE distribution data found")
		return nil, nil
	}
	return results, nil
}

func (d *ClickHouseQueryDriver) GetQPM(periodInSeconds uint64, limit uint16) ([]models.TimeSeriesData, error) {
	cutoffTime := time.Now().Add(-time.Duration(periodInSeconds) * time.Second)

	rows, err := d.clickHouseDB.Query(context.Background(), `
		SELECT
			toStartOfMinute(timestamp) as time,
			count(*) as requests,
			sum(1 - success) as errors,
			(countIf(success = 1) * 100.0) / count(*) as percentage
		FROM dns_metrics
		WHERE timestamp >= ?
		GROUP BY time
		ORDER BY time DESC
		LIMIT ?;
	`, cutoffTime, limit)
	if err != nil {
		d.logger.Error("Failed to query QPM data", "error", err)
		return nil, fmt.Errorf("failed to query QPM data: %w", err)
	}
	defer rows.Close()

	var results []models.TimeSeriesData
	for rows.Next() {
		var data models.TimeSeriesData
		if err := rows.Scan(&data.Time, &data.Requests, &data.Errors, &data.Percentage); err != nil {
			d.logger.Error("Failed to scan QPM row", "error", err)
			return nil, fmt.Errorf("failed to scan QPM row: %w", err)
		}
		if math.IsNaN(data.Percentage) {
			data.Percentage = 0
		}
		results = append(results, data)
	}

	if err := rows.Err(); err != nil {
		d.logger.Error("Error iterating over QPM rows", "error", err)
		return nil, fmt.Errorf("error iterating over QPM rows: %w", err)
	}

	if len(results) == 0 {
		d.logger.Info("No QPM data found")
		return nil, nil
	}

	return results, nil
}
