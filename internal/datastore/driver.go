package datastore

import (
	"github.com/Unfield/Odin-DNS/internal/types"
)

type Driver interface {
	GetUser(id string) (*types.User, error)
	GetUserById(id string) (*types.User, error)
	CreateUser(user *types.User) error
	UpdateUser(user *types.User) error

	GetSession(id string) (*types.Session, error)
	CreateSession(session *types.Session) error
	UpdateSession(session *types.Session) error

	CreateZone(zone *types.DBZone) error
	UpdateZone(zone *types.DBZone) error
	GetRecord(id string) (*types.DBRecord, error)
	GetRecordByName(name string) (*types.DBRecord, error)
	CreateRecord(record *types.DBRecord) error
	UpdateRecord(record *types.DBRecord) error
	GetFullZone(name string) (*types.DBZone, []types.DBRecord, error)
	GetFullZoneById(id string) (*types.DBZone, []types.DBRecord, error)
}
