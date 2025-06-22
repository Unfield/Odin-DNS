package mysql

import (
	"fmt"

	"github.com/Unfield/Odin-DNS/internal/types"
)

func (d *MySQLDriver) GetUser(username string) (*types.User, error) {
	query := "SELECT id, username, password_hash, created_at, updated_at, deleted_at FROM users WHERE username = ?"
	var user types.User
	err := d.db.Get(&user, query, username)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			d.logger.Info("User not found", "username", username)
			return nil, nil
		}
		d.logger.Error("Failed to get user", "error", err)
		return nil, err
	}
	return &user, nil
}

func (d *MySQLDriver) GetUserById(id string) (*types.User, error) {
	query := "SELECT id, username, password_hash, created_at, updated_at, deleted_at FROM users WHERE id = ?"
	var user types.User
	err := d.db.Get(&user, query, id)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			d.logger.Info("User not found", "id", id)
			return nil, nil
		}
		d.logger.Error("Failed to get user", "error", err)
		return nil, err
	}
	return &user, nil
}

func (d *MySQLDriver) CreateUser(user *types.User) error {
	query := "INSERT INTO users (id, username, password_hash, created_at, updated_at) VALUES (?, ?, ?, NOW(), NOW())"
	_, err := d.db.Exec(query, user.ID, user.Username, user.PasswordHash)
	if err != nil {
		d.logger.Error("Failed to create user", "error", err)
		if err.Error() == "UNIQUE constraint failed: users.username" {
			d.logger.Info("Username already exists", "username", user.Username)
			return fmt.Errorf("username already exists: %s", user.Username)
		}
		d.logger.Error("Failed to create user", "error", err)
	}
	return err
}

func (d *MySQLDriver) UpdateUser(user *types.User) error {
	return nil
}

func (d *MySQLDriver) GetSession(id string) (*types.Session, error) {
	query := "SELECT id, user_id, token, created_at, updated_at, deleted_at FROM sessions WHERE id = ?"
	var session types.Session
	err := d.db.Get(&session, query, id)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			d.logger.Info("Session not found", "id", id)
			return nil, nil
		}
		d.logger.Error("Failed to get session", "error", err)
		return nil, err
	}
	return &session, nil
}

func (d *MySQLDriver) CreateSession(session *types.Session) error {
	query := "INSERT INTO sessions (id, user_id, token, created_at, updated_at) VALUES (?, ?, ?, NOW(), NOW())"
	_, err := d.db.Exec(query, session.ID, session.UserID, session.Token)
	if err != nil {
		d.logger.Error("Failed to create session", "error", err)
		return err
	}
	return nil
}

func (d *MySQLDriver) UpdateSession(session *types.Session) error {
	query := "UPDATE sessions SET token = ?, updated_at = NOW(), deleted_at = ? WHERE id = ?"
	_, err := d.db.Exec(query, session.Token, session.DeletedAt, session.ID)
	if err != nil {
		d.logger.Error("Failed to update session", "error", err)
		return err
	}
	return nil
}
