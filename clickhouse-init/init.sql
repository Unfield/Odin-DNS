CREATE DATABASE IF NOT EXISTS default;

CREATE TABLE IF NOT EXISTS default.dns_metrics (
    timestamp DateTime ('UTC'),
    ip String,
    domain String,
    query_type String,
    success UInt8,
    error_message String,
    response_time_ms Float64,
    cache_hit UInt8,
    rcode UInt8
) ENGINE = MergeTree ()
PARTITION BY
    toYYYYMM (timestamp)
ORDER BY
    (timestamp, domain, ip) TTL timestamp + INTERVAL 190 DAY SETTINGS index_granularity = 8192;
