package datastore

import (
	"github.com/Unfield/Odin-DNS/internal/types"
	"github.com/Unfield/Odin-DNS/pkg/odintypes"
)

type Driver interface {
	GetUser(id string) (*types.User, error)
	GetUserById(id string) (*types.User, error)
	CreateUser(user *types.User) error
	UpdateUser(user *types.User) error

	GetSession(id string) (*types.Session, error)
	GetSessionByToken(token string) (*types.Session, error)
	CreateSession(session *types.Session) error
	UpdateSession(session *types.Session) error

	GetZone(id string) (*types.DBZone, error)
	CreateZone(zone *types.DBZone) error
	UpdateZone(zone *types.DBZone) error
	DeleteZone(id string) error
	GetRecord(id string) (*types.DBRecord, error)
	GetRecordByName(name string) (*types.DBRecord, error)
	CreateRecord(record *types.DBRecord) error
	UpdateRecord(record *types.DBRecord) error
	DeleteRecord(id string) error
	GetFullZone(name string) (*types.DBZone, []types.DBRecord, error)
	GetFullZoneById(id string) (*types.DBZone, []types.DBRecord, error)

	GetZones(owner string) ([]types.DBZone, error)
	GetZoneEntries(zoneId string) ([]types.DBRecord, error)

	LookupRecordForDNSQuery(rname string, rtype uint16, rclass uint16) (*odintypes.DNSRecord, uint8, error)
}
