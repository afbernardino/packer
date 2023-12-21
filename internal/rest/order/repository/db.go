package repository

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/jackc/pgx/v5/pgtype"
)

// Database can communicate with the persistent repository.
type Database struct {
	handler *sql.DB
}

func NewDatabase(handler *sql.DB) Database {
	return Database{handler: handler}
}

func (db *Database) SetConfig(ctx context.Context, cfg Config) error {
	stmt, err := db.handler.PrepareContext(ctx, `UPDATE orders_config SET pack_sizes = $1`)
	if err != nil {
		return fmt.Errorf("error preparing statment to set the config: %w", err)
	}

	_, err = stmt.ExecContext(ctx, cfg.PackSizes)
	if err != nil {
		return fmt.Errorf("error updating config: %w", err)
	}

	return nil
}

func (db *Database) FindConfig(ctx context.Context) (Config, error) {
	var cfg Config
	m := pgtype.NewMap()
	err := db.handler.QueryRowContext(ctx, "SELECT pack_sizes FROM orders_config").Scan(m.SQLScanner(&cfg.PackSizes))
	if err != nil {
		return Config{}, fmt.Errorf("error querying config: %w", err)
	}
	return cfg, nil
}
