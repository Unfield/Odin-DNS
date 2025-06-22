package mysql

import (
	"github.com/Unfield/Odin-DNS/internal/types"
)

func (d *MySQLDriver) CreateZone(zone *types.DBZone) (err error) {
	query := "INSERT INTO zones (id, owner, name, created_at, updated_at) VALUES (?, ?, ?, ?, ?)"

	_, err = d.db.Exec(query, zone.ID, zone.Owner, zone.Name, zone.CreatedAt, zone.UpdatedAt)
	if err != nil {
		d.logger.Error("Failed to create zone", "error", err)
		return err
	}
	return nil
}

func (d *MySQLDriver) UpdateZone(zone *types.DBZone) error {
	query := "UPDATE zones SET name = ?, updated_at = ?, deleted_at = ? WHERE id = ?"
	_, err := d.db.Exec(query, zone.Name, zone.UpdatedAt, zone.DeletedAt, zone.ID)
	if err != nil {
		d.logger.Error("Failed to update zone", "error", err)
		return err
	}
	return nil
}

func (d *MySQLDriver) GetFullZone(name string) (*types.DBZone, []types.DBRecord, error) {
	query := "SELECT id, owner, name, created_at, updated_at FROM zones WHERE name = ?"
	var zone types.DBZone
	err := d.db.Get(&zone, query, name)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			d.logger.Info("Zone not found", "name", name)
			return nil, nil, nil
		}
		d.logger.Error("Failed to get zone", "error", err)
		return nil, nil, err
	}

	recordQuery := "SELECT id, zone_id, name, type, class, ttl, rdata, created_at, updated_at FROM zone_entries WHERE zone_id = ?"
	var records []types.DBRecord
	err = d.db.Select(&records, recordQuery, zone.ID)
	if err != nil {
		d.logger.Error("Failed to get records for zone", "error", err)
		return nil, nil, err
	}

	return &zone, records, nil
}

func (d *MySQLDriver) GetFullZoneById(id string) (*types.DBZone, []types.DBRecord, error) {
	query := "SELECT id, owner, name, created_at, updated_at FROM zones WHERE id = ?"
	var zone types.DBZone
	err := d.db.Get(&zone, query, id)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			d.logger.Info("Zone not found", "id", id)
			return nil, nil, nil
		}
		d.logger.Error("Failed to get zone by ID", "error", err)
		return nil, nil, err
	}

	recordQuery := "SELECT id, zone_id, name, type, class, ttl, rdata, created_at, updated_at FROM zone_entries WHERE zone_id = ?"
	var records []types.DBRecord
	err = d.db.Select(&records, recordQuery, zone.ID)
	if err != nil {
		d.logger.Error("Failed to get records for zone by ID", "error", err)
		return nil, nil, err
	}

	return &zone, records, nil
}

func (d *MySQLDriver) GetRecord(id string) (*types.DBRecord, error) {
	query := "SELECT id, zone_id, name, type, class, ttl, rdata, created_at, updated_at FROM zone_entries WHERE id = ?"
	var record types.DBRecord
	err := d.db.Get(&record, query, id)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			d.logger.Info("Record not found", "id", id)
			return nil, nil
		}
		d.logger.Error("Failed to get record", "error", err)
		return nil, err
	}
	return &record, nil
}

func (d *MySQLDriver) GetRecordByName(name string) (*types.DBRecord, error) {
	query := "SELECT id, zone_id, name, type, class, ttl, rdata, created_at, updated_at FROM zone_entries WHERE name = ?"
	var record types.DBRecord
	err := d.db.Get(&record, query, name)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			d.logger.Info("Record not found", "name", name)
			return nil, nil
		}
		d.logger.Error("Failed to get record by name", "error", err)
		return nil, err
	}
	return &record, nil
}

func (d *MySQLDriver) CreateRecord(record *types.DBRecord) error {
	query := "INSERT INTO zone_entries (id, zone_id, name, type, class, ttl, rdata, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, NOW(), NOW())"
	_, err := d.db.Exec(query, record.ID, record.ZoneID, record.Name, record.Type, record.Class, record.TTL, record.RData)
	if err != nil {
		d.logger.Error("Failed to create record", "error", err)
		return err
	}
	return nil
}

func (d *MySQLDriver) UpdateRecord(record *types.DBRecord) error {
	query := "UPDATE zone_entries SET zone_id = ?, name = ?, type = ?, class = ?, ttl = ?, rdata = ?, updated_at = NOW() WHERE id = ?"
	_, err := d.db.Exec(query, record.ZoneID, record.Name, record.Type, record.Class, record.TTL, record.RData, record.ID)
	if err != nil {
		d.logger.Error("Failed to update record", "error", err)
		return err
	}
	return nil
}
