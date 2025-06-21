package types

import (
	"database/sql"
	"time"
)

type User struct {
	ID           string       `json:"id" db:"id"`
	Username     string       `json:"username" db:"username"`
	PasswordHash string       `json:"password_hash" db:"password_hash"`
	CreatedAt    time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at" db:"updated_at"`
	DeletedAt    sql.NullTime `json:"deleted_at" db:"deleted_at"`
}

type Session struct {
	ID        string       `json:"id" db:"id"`
	UserID    string       `json:"user_id" db:"user_id"`
	Token     string       `json:"token" db:"token"`
	CreatedAt time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt time.Time    `json:"updated_at" db:"updated_at"`
	DeletedAt sql.NullTime `json:"deleted_at" db:"deleted_at"`
}
