package metrics

import (
	"context"
	"log/slog"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/Unfield/Odin-DNS/internal/config"
)

type ClickHouseIngestionDriver struct {
	clickHouseDB  driver.Conn
	metricBuffer  chan DNSMetric
	logger        *slog.Logger
	batchSize     int
	batchInterval time.Duration
}

func NewClickHouseIngestionDriver(config *config.Config) MetricsIngestionDriver {
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

	driver := &ClickHouseIngestionDriver{
		clickHouseDB:  conn,
		metricBuffer:  make(chan DNSMetric, config.CLICKHOUSE_MAX_BATCH_SIZE*2),
		logger:        slog.Default().WithGroup("Metrics"),
		batchSize:     config.CLICKHOUSE_MAX_BATCH_SIZE,
		batchInterval: config.CLICKHOUSE_BATCH_INTERVAL,
	}
	go driver.ProcessMetricsBatch()
	return driver
}

func (d *ClickHouseIngestionDriver) Close() error {
	if d.clickHouseDB != nil {
		d.logger.Info("Attempting to flush remaining metrics before closing ClickHouse connection...")
		if len(d.metricBuffer) > 0 {
			remainingBatch := make([]DNSMetric, 0, len(d.metricBuffer))
			for len(d.metricBuffer) > 0 {
				remainingBatch = append(remainingBatch, <-d.metricBuffer)
			}
			if len(remainingBatch) > 0 {
				if err := d.writeBatch(remainingBatch); err != nil {
					d.logger.Error("Failed to write remaining batch to ClickHouse during shutdown", "error", err)
				} else {
					d.logger.Info("Successfully flushed remaining metrics during shutdown.")
				}
			}
		}

		if err := d.clickHouseDB.Close(); err != nil {
			d.logger.Error("Failed to close ClickHouse connection", "error", err)
			return err
		}
	}
	d.logger.Info("ClickHouse connection closed")
	return nil
}

func (d *ClickHouseIngestionDriver) Collect(metric DNSMetric) {
	select {
	case d.metricBuffer <- metric:
	default:
		d.logger.Warn("Metric buffer is full, dropping metric", "metric", metric)
	}
}

func (d *ClickHouseIngestionDriver) ProcessMetricsBatch() {
	ticker := time.NewTicker(d.batchInterval)
	defer ticker.Stop()

	var batch []DNSMetric

	for {
		select {
		case metric := <-d.metricBuffer:
			batch = append(batch, metric)
			if len(batch) >= d.batchSize {
				if err := d.writeBatch(batch); err != nil {
					d.logger.Error("Failed to write batch to ClickHouse", "error", err)
				}
				batch = nil
			}
		case <-ticker.C:
			if len(batch) > 0 {
				if err := d.writeBatch(batch); err != nil {
					d.logger.Error("Failed to write batch to ClickHouse", "error", err)
				}
				batch = nil
			}
		}
	}
}

func (d *ClickHouseIngestionDriver) writeBatch(batch []DNSMetric) error {
	if len(batch) == 0 {
		return nil
	}

	batchWriter, err := d.clickHouseDB.PrepareBatch(context.Background(), "INSERT INTO dns_metrics (timestamp, ip, domain, query_type, success, error_message, response_time_ms, cache_hit, rcode)")
	if err != nil {
		return err
	}
	defer batchWriter.Send()

	for _, m := range batch {
		err = batchWriter.Append(
			m.Timestamp,
			m.IP,
			m.Domain,
			m.QueryType,
			m.Success,
			m.ErrorMessage,
			m.ResponseTimeMs,
			m.CacheHit,
			m.Rcode,
		)
		if err != nil {
			d.logger.Error("Failed to append metric to batch", "metric", m, "error", err)
		}
	}

	return batchWriter.Send()
}
