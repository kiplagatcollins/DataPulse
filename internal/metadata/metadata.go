// Package metadata orchestrates database metadata lookups for the explorer.
package metadata

import (
	"context"
	"fmt"

	"github.com/collins/GoDBClient/internal/connpool"
	"github.com/collins/GoDBClient/internal/models"
)

// Manager resolves metadata for a live connection.
type Manager struct {
	pool *connpool.Manager
}

// NewManager returns a metadata manager bound to a connection pool.
func NewManager(p *connpool.Manager) *Manager {
	return &Manager{pool: p}
}

// Databases lists the databases or schemas visible to the connection.
func (m *Manager) Databases(ctx context.Context, connID string) ([]string, error) {
	conn, err := m.pool.Get(connID)
	if err != nil {
		return nil, fmt.Errorf("metadata: %w", err)
	}
	catalog := conn.Catalog()
	if catalog == nil {
		return nil, fmt.Errorf("metadata: connection does not support catalogs")
	}
	return catalog.Databases(ctx)
}

// Tables lists tables and views for a schema.
func (m *Manager) Tables(ctx context.Context, connID string, schema string) ([]models.TableInfo, error) {
	conn, err := m.pool.Get(connID)
	if err != nil {
		return nil, fmt.Errorf("metadata: %w", err)
	}
	catalog := conn.Catalog()
	if catalog == nil {
		return nil, fmt.Errorf("metadata: connection does not support catalogs")
	}
	return catalog.Tables(ctx, schema)
}

// Columns lists columns for a table.
func (m *Manager) Columns(ctx context.Context, connID string, schema string, table string) ([]models.ColumnInfo, error) {
	conn, err := m.pool.Get(connID)
	if err != nil {
		return nil, fmt.Errorf("metadata: %w", err)
	}
	catalog := conn.Catalog()
	if catalog == nil {
		return nil, fmt.Errorf("metadata: connection does not support catalogs")
	}
	return catalog.Columns(ctx, schema, table)
}

// ForeignKeys lists foreign keys for a table.
func (m *Manager) ForeignKeys(ctx context.Context, connID string, schema string, table string) ([]models.ForeignKeyInfo, error) {
	conn, err := m.pool.Get(connID)
	if err != nil {
		return nil, fmt.Errorf("metadata: %w", err)
	}
	catalog := conn.Catalog()
	if catalog == nil {
		return nil, fmt.Errorf("metadata: connection does not support catalogs")
	}
	return catalog.ForeignKeys(ctx, schema, table)
}

// Indexes lists indexes for a table.
func (m *Manager) Indexes(ctx context.Context, connID string, schema string, table string) ([]string, error) {
	conn, err := m.pool.Get(connID)
	if err != nil {
		return nil, fmt.Errorf("metadata: %w", err)
	}
	catalog := conn.Catalog()
	if catalog == nil {
		return nil, fmt.Errorf("metadata: connection does not support catalogs")
	}
	return catalog.Indexes(ctx, schema, table)
}
