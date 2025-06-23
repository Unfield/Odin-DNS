package redis

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/Unfield/Odin-DNS/internal/datastore"
	"github.com/Unfield/Odin-DNS/internal/types"
	"github.com/Unfield/Odin-DNS/internal/util"
	"github.com/Unfield/Odin-DNS/pkg/odintypes"
	"github.com/redis/go-redis/v9"
)

type RedisCacheDriver struct {
	redisClient *redis.Client
	datastore.Driver
	logger  *slog.Logger
	context context.Context
}

func NewRedisCacheDriver(persistentDriver datastore.Driver, addr, username, password string, db int) *RedisCacheDriver {
	return &RedisCacheDriver{
		redisClient: redis.NewClient(&redis.Options{
			Addr:     addr,
			Username: username,
			Password: password,
			DB:       db,
			TLSConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
			},
		}),
		Driver:  persistentDriver,
		logger:  slog.Default().WithGroup("Redis-Driver"),
		context: context.Background(),
	}
}

func (d *RedisCacheDriver) Close() error {
	return d.redisClient.Close()
}

func (d *RedisCacheDriver) LookupRecordForDNSQuery(rname string, rtype uint16, rclass uint16) (*odintypes.DNSRecord, error) {
	rTypeStr := odintypes.TypeToString(rtype)
	rClassStr := odintypes.ClassToString(rclass)
	cacheKey := combineSearchPartsToKey(rname, rtype, rclass)

	cacheEntry, err := d.redisClient.Get(d.context, cacheKey).Result()
	if err != nil {
		if err == redis.Nil {
			d.logger.Info("Cache miss", "name", rname, "type", rTypeStr, "class", rClassStr)
			dbRecordFromPersistent, err := d.Driver.LookupRecordForDNSQuery(rname, rtype, rclass)
			if err != nil {
				return nil, err
			}
			if dbRecordFromPersistent == nil {
				d.logger.Info("Record not found in persistent store", "name", rname)
				return nil, nil
			}

			rDataStringForCache := util.ConvertRDataBytesToString(dbRecordFromPersistent.Type, dbRecordFromPersistent.RData)
			if rDataStringForCache == "" && len(dbRecordFromPersistent.RData) > 0 {
				d.logger.Warn("Failed to convert RData bytes to string for caching; not caching this RData.",
					"type", dbRecordFromPersistent.Type, "rname", rname)
			}

			cacheableRecord := types.CacheRecord{
				Name:  dbRecordFromPersistent.Name,
				Type:  odintypes.TypeToString(dbRecordFromPersistent.Type),
				Class: odintypes.ClassToString(dbRecordFromPersistent.Class),
				TTL:   dbRecordFromPersistent.TTL,
				RData: rDataStringForCache,
			}

			recordJSONBytes, marshalErr := json.Marshal(cacheableRecord)
			if marshalErr != nil {
				d.logger.Error("Failed to marshal DNS record for caching", "error", marshalErr, "record", cacheableRecord)
			} else {
				cacheTTL := time.Duration(dbRecordFromPersistent.TTL) * time.Second
				if cacheTTL <= 0 {
					cacheTTL = 5 * time.Minute
				}

				if setErr := d.redisClient.Set(d.context, cacheKey, recordJSONBytes, cacheTTL).Err(); setErr != nil {
					d.logger.Error("Failed to set DNS record in cache", "error", setErr, "key", cacheKey)
				} else {
					d.logger.Info("Record cached successfully", "name", rname, "type", rTypeStr, "class", rClassStr, "ttl", cacheTTL)
				}
			}
			return dbRecordFromPersistent, nil
		} else {
			d.logger.Error("Failed to retrieve data from cache", "error", err, "key", cacheKey)
			return nil, fmt.Errorf("cache query failed for %s (%s, %s): %w", rname, rTypeStr, rClassStr, err)
		}
	}

	var cachedDBRecord types.CacheRecord
	if err := json.Unmarshal([]byte(cacheEntry), &cachedDBRecord); err != nil {
		d.logger.Error("Failed to unmarshal DNS record from cache (corrupted?)", "error", err, "cache_entry", cacheEntry)
		d.redisClient.Del(d.context, cacheKey)
		d.logger.Info("Attempting to fetch from persistent store after unmarshal error", "name", rname)
		return d.Driver.LookupRecordForDNSQuery(rname, rtype, rclass)
	}

	packedRData, convErr := util.ConvertRDataStringToBytes(rtype, cachedDBRecord.RData)
	if convErr != nil {
		d.logger.Error("Failed to convert RData string to bytes from cache entry (corrupted?)",
			"type", cachedDBRecord.Type, "rdata_string", cachedDBRecord.RData, "error", convErr)
		d.redisClient.Del(d.context, cacheKey)
		d.logger.Info("Attempting to fetch from persistent store after RData conversion error", "name", rname)
		return d.Driver.LookupRecordForDNSQuery(rname, rtype, rclass)
	}

	d.logger.Info("Cache hit", "name", rname, "type", rTypeStr, "class", rClassStr)

	return &odintypes.DNSRecord{
		Name:  cachedDBRecord.Name,
		Type:  rtype,
		Class: rclass,
		TTL:   cachedDBRecord.TTL,
		RData: packedRData,
	}, nil
}

func combineSearchPartsToKey(rname string, rtype uint16, rclass uint16) string {
	return fmt.Sprintf("%s|%d|%d", rname, rtype, rclass)
}

func (d *RedisCacheDriver) CreateRecord(record *types.DBRecord) error {
	d.logger.Info("Creating record in persistent store",
		"name", record.Name, "type", record.Type, "class", record.Class)
	if err := d.Driver.CreateRecord(record); err != nil {
		return fmt.Errorf("failed to create record in persistent store: %w", err)
	}

	recordJSONBytes, err := json.Marshal(record)
	if err != nil {
		d.logger.Error("Failed to marshal record for caching during creation", "error", err, "record", record)
		return nil
	}

	recordTypeUint, parseTypeErr := odintypes.StringToType(record.Type)
	recordClassUint, parseClassErr := odintypes.StringToClass(record.Class)

	if parseTypeErr != nil || parseClassErr != nil {
		d.logger.Error("Failed to parse record type/class for cache key during creation",
			"type_str", record.Type, "class_str", record.Class,
			"type_err", parseTypeErr, "class_err", parseClassErr)
		return nil
	}

	cacheKey := combineSearchPartsToKey(record.Name, recordTypeUint, recordClassUint)
	cacheTTL := time.Duration(record.TTL) * time.Second
	if cacheTTL <= 0 {
		cacheTTL = 5 * time.Minute
	}

	if setErr := d.redisClient.Set(d.context, cacheKey, recordJSONBytes, cacheTTL).Err(); setErr != nil {
		d.logger.Error("Failed to cache record during creation", "error", setErr, "key", cacheKey)
		return nil
	}

	d.logger.Info("Record successfully created and cached",
		"name", record.Name, "type", record.Type, "class", record.Class, "ttl", cacheTTL)

	return nil
}
