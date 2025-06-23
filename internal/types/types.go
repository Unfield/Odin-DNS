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

type DBZone struct {
	ID        string       `json:"id" db:"id"`
	Owner     string       `json:"owner" db:"owner"`
	Name      string       `json:"name" db:"name"`
	CreatedAt time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt time.Time    `json:"updated_at" db:"updated_at"`
	DeletedAt sql.NullTime `json:"deleted_at" db:"deleted_at"`
}

type DBRecord struct {
	ID        string       `json:"id" db:"id"`
	ZoneID    string       `json:"zone_id" db:"zone_id"`
	Name      string       `json:"name" db:"name"`
	Type      string       `json:"type" db:"type"`
	Class     string       `json:"class" db:"class"`
	TTL       uint32       `json:"ttl" db:"ttl"`
	RData     string       `json:"rdata" db:"rdata"`
	CreatedAt time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt time.Time    `json:"updated_at" db:"updated_at"`
	DeletedAt sql.NullTime `json:"deleted_at" db:"deleted_at"`
}

type CacheRecord struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Class string `json:"class"`
	TTL   uint32 `json:"ttl"`
	RData string `json:"rdata"`
}
