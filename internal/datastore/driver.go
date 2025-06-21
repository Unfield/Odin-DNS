package datastore

import "github.com/Unfield/Odin-DNS/internal/types"

type Driver interface {
	GetUser(id string) (*types.User, error)
	GetUserById(id string) (*types.User, error)
	CreateUser(user *types.User) error
	UpdateUser(user *types.User) error

	GetSession(id string) (*types.Session, error)
	CreateSession(session *types.Session) error
	UpdateSession(session *types.Session) error
}
