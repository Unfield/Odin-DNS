package mysql

import (
	"log/slog"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type MySQLDriver struct {
	db     *sqlx.DB
	logger *slog.Logger
}

func NewMySQLDriver(dsn string) (*MySQLDriver, error) {
	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &MySQLDriver{
		db:     db,
		logger: slog.Default().WithGroup("MySQL-Driver"),
	}, nil
}

func (d *MySQLDriver) Close() error {
	return d.db.Close()
}
