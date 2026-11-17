// Package driver defines the database driver abstraction used by DataPulse.
package driver

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/collins/DataPulse/internal/models"
)

// Type is the canonical database family identifier.
type Type string

const (
	Postgres Type = "postgres"
	MySQL    Type = "mysql"
	SQLite   Type = "sqlite"
	Oracle   Type = "oracle"
	Redis    Type = "redis"
)

// Catalog provides access to database metadata for the explorer tree.
type Catalog interface {
	Databases(ctx context.Context) ([]string, error)
	Tables(ctx context.Context, schema string) ([]models.TableInfo, error)
	Columns(ctx context.Context, schema string, table string) ([]models.ColumnInfo, error)
	ForeignKeys(ctx context.Context, schema string, table string) ([]models.ForeignKeyInfo, error)
	Indexes(ctx context.Context, schema string, table string) ([]string, error)
}

// KV provides key-value access for key-value stores behind the same Conn.
type KV interface {
	ScanKeys(ctx context.Context, pattern string, cursor uint64, count int64) (keys []string, next uint64, err error)
	KeyInfo(ctx context.Context, key string) (models.KeyInfo, error)
	GetValue(ctx context.Context, key string) (models.ValueView, error)
	Exec(ctx context.Context, args []string) ([]string, error)
}

// Conn is the runtime handle for a live database connection.
type Conn interface {
	Close() error
	Catalog() Catalog
	KV() KV
}

// SQLExec is implemented by database/sql backed connections.
type SQLExec interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

// Driver opens and provides access to a database.
type Driver interface {
	Type() Type
	Open(ctx context.Context, profile *models.Profile) (Conn, error)
}

// Registry maps driver types to implementations.
type Registry struct {
	drivers map[Type]Driver
}

// NewRegistry returns an empty driver registry.
func NewRegistry() *Registry {
	return &Registry{drivers: map[Type]Driver{}}
}

// Register adds a driver to the registry.
func (r *Registry) Register(d Driver) {
	if d == nil {
		return
	}
	r.drivers[d.Type()] = d
}

// Get returns the driver registered for the given type.
func (r *Registry) Get(t Type) (Driver, error) {
	d, ok := r.drivers[t]
	if !ok {
		return nil, fmt.Errorf("driver not registered for type %q", t)
	}
	return d, nil
}
