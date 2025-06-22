package mysql

import (
	"database/sql"
	"fmt"

	"github.com/Unfield/Odin-DNS/internal/util"
	"github.com/Unfield/Odin-DNS/pkg/odintypes"
)

type DBRecord struct {
	Name  string
	Type  string
	Class string
	TTL   uint32
	RData string
}

func (d *MySQLDriver) LookupRecordForDNSQuery(rname string, rtype uint16, rclass uint16) (*odintypes.DNSRecord, error) {
	query := "SELECT name, type, class, ttl, rdata FROM zone_entries WHERE name = ? AND type = ? AND class = ?"

	var dbRecord DBRecord

	rTypeStr := odintypes.TypeToString(rtype)
	rClassStr := odintypes.ClassToString(rclass)

	d.logger.Debug("Attempting DB Get", "name", rname, "type_str", rTypeStr, "class_str", rClassStr)

	err := d.db.Get(&dbRecord, query, rname, rTypeStr, rClassStr)

	if err != nil {
		if err == sql.ErrNoRows {
			d.logger.Debug("Record not found in DB (sql.ErrNoRows)", "name", rname, "type", rTypeStr, "class", rClassStr)
			return nil, nil
		}
		d.logger.Error("Failed to scan record from DB or other SQL error", "error", err, "name", rname, "type", rTypeStr, "class", rClassStr)
		return nil, fmt.Errorf("database query failed for %s (%s, %s): %w", rname, rTypeStr, rClassStr, err)
	}

	packedRData, convErr := util.ConvertRDataStringToBytes(rtype, dbRecord.RData)
	if convErr != nil {
		d.logger.Error("Failed to convert RData string to bytes", "type", dbRecord.Type, "rdata_string", dbRecord.RData, "error", convErr)
		return nil, fmt.Errorf("failed to convert RData string '%s' for type %s: %w", dbRecord.RData, dbRecord.Type, convErr)
	}
	d.logger.Info("RData successfully converted", "packed_len", len(packedRData))

	return &odintypes.DNSRecord{
		Name:  dbRecord.Name,
		Type:  rtype,
		Class: rclass,
		TTL:   dbRecord.TTL,
		RData: packedRData,
	}, nil
}
