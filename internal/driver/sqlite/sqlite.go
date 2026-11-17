// Package sqlite implements the SQLite driver for DataPulse.
package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/collins/DataPulse/internal/driver"
	"github.com/collins/DataPulse/internal/models"
	_ "modernc.org/sqlite"
)

type Driver struct{}

// NewDriver returns a SQLite driver.
func NewDriver() *Driver {
	return &Driver{}
}

// Type returns the SQLite driver type.
func (d *Driver) Type() driver.Type {
	return driver.SQLite
}

// Open opens a SQLite connection from a profile.
func (d *Driver) Open(ctx context.Context, profile *models.Profile) (driver.Conn, error) {
	path := profile.Database
	if path == "" {
		path = ":memory:"
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("sqlite: open connection: %w", err)
	}
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("sqlite: ping connection: %w", err)
	}
	return &sqliteConn{db: db}, nil
}

// sqliteConn adapts a database/sql database to the driver connection port.
type sqliteConn struct {
	db *sql.DB
}

// Close closes the underlying database connection.
func (c *sqliteConn) Close() error {
	return c.db.Close()
}

// Catalog returns the metadata catalog for the connection.
func (c *sqliteConn) Catalog() driver.Catalog {
	return &sqliteCatalog{db: c.db}
}

// KV returns nil because SQLite is not a key-value store.
func (c *sqliteConn) KV() driver.KV {
	return nil
}

// QueryContext executes a query that returns rows.
func (c *sqliteConn) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return c.db.QueryContext(ctx, query, args...)
}

// ExecContext executes a statement that does not return rows.
func (c *sqliteConn) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return c.db.ExecContext(ctx, query, args...)
}

// sqliteCatalog implements driver.Catalog for SQLite.
type sqliteCatalog struct {
	db *sql.DB
}

// Databases lists databases visible to the connection.
func (c *sqliteCatalog) Databases(ctx context.Context) ([]string, error) {
	return []string{"main"}, nil
}

// Tables lists tables and views for a schema.
func (c *sqliteCatalog) Tables(ctx context.Context, schema string) ([]models.TableInfo, error) {
	rows, err := c.db.QueryContext(ctx, "SELECT name, type FROM sqlite_master WHERE type IN ('table', 'view') AND name NOT LIKE 'sqlite_%' ORDER BY name")
	if err != nil {
		return nil, fmt.Errorf("sqlite: list tables: %w", err)
	}
	defer rows.Close()
	var out []models.TableInfo
	for rows.Next() {
		var name, tableType string
		if err := rows.Scan(&name, &tableType); err != nil {
			return nil, fmt.Errorf("sqlite: scan table: %w", err)
		}
		out = append(out, models.TableInfo{
			Schema: "main",
			Name:   name,
			Type:   tableType,
		})
	}
	return out, rows.Err()
}

// Columns lists columns for a table.
func (c *sqliteCatalog) Columns(ctx context.Context, schema string, table string) ([]models.ColumnInfo, error) {
	query := fmt.Sprintf("SELECT cid, name, type, \"notnull\", dflt_value, pk FROM pragma_table_info('%s') ORDER BY cid", table)
	rows, err := c.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("sqlite: list columns: %w", err)
	}
	defer rows.Close()
	var out []models.ColumnInfo
	for rows.Next() {
		var cid int
		var name, dataType string
		var notNull, pk int64
		var defaultValue any
		if err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk); err != nil {
			return nil, fmt.Errorf("sqlite: scan column: %w", err)
		}
		out = append(out, models.ColumnInfo{
			Name:      name,
			Type:      dataType,
			NotNull:   notNull == 1,
			Nullable:  notNull == 0,
			Default:   defaultValue,
			IsPrimary: pk == 1,
		})
	}
	return out, rows.Err()
}

// ForeignKeys lists foreign keys for a table.
func (c *sqliteCatalog) ForeignKeys(ctx context.Context, schema string, table string) ([]models.ForeignKeyInfo, error) {
	rows, err := c.db.QueryContext(ctx, fmt.Sprintf("SELECT id, seq, \"table\", \"from\", \"to\" FROM pragma_foreign_key_list('%s') ORDER BY seq", table))
	if err != nil {
		return nil, fmt.Errorf("sqlite: list foreign keys: %w", err)
	}
	defer rows.Close()
	var out []models.ForeignKeyInfo
	for rows.Next() {
		var id, seq int
		var refTable, from, to string
		if err := rows.Scan(&id, &seq, &refTable, &from, &to); err != nil {
			return nil, fmt.Errorf("sqlite: scan foreign key: %w", err)
		}
		out = append(out, models.ForeignKeyInfo{
			Name:       from,
			Columns:    []string{from},
			RefTable:   refTable,
			RefColumns: []string{to},
		})
	}
	return out, rows.Err()
}

// Indexes lists indexes for a table.
func (c *sqliteCatalog) Indexes(ctx context.Context, schema string, table string) ([]string, error) {
	rows, err := c.db.QueryContext(ctx, "SELECT name FROM sqlite_master WHERE type = 'index' AND tbl_name = ? ORDER BY name", table)
	if err != nil {
		return nil, fmt.Errorf("sqlite: list indexes: %w", err)
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("sqlite: scan index: %w", err)
		}
		out = append(out, name)
	}
	return out, rows.Err()
}
