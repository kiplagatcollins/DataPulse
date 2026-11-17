// Package oracle implements the Oracle driver for GoDBClient.
package oracle

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/collins/GoDBClient/internal/driver"
	"github.com/collins/GoDBClient/internal/models"
	go_ora "github.com/sijms/go-ora/v2"
	"github.com/zalando/go-keyring"
)

const (
	serviceName = "GoDBClient"
	defaultHost = "localhost"
	defaultPort = 1521
	defaultSvc  = "xe"
)

// Driver opens Oracle connections.
type Driver struct{}

// NewDriver returns an Oracle driver.
func NewDriver() *Driver {
	return &Driver{}
}

// Type returns the Oracle driver type.
func (d *Driver) Type() driver.Type {
	return driver.Oracle
}

// Open opens an Oracle connection from a profile.
func (d *Driver) Open(ctx context.Context, profile *models.Profile) (driver.Conn, error) {
	password, err := keyring.Get(serviceName, profile.ID)
	if err != nil {
		return nil, fmt.Errorf("oracle: read keychain secret: %w", err)
	}
	host := profile.Host
	if host == "" {
		host = defaultHost
	}
	port := profile.Port
	if port == 0 {
		port = defaultPort
	}
	service := profile.Database
	if service == "" {
		service = defaultSvc
	}
	user := profile.Username
	if user == "" {
		user = "system"
	}
	dsn := fmt.Sprintf("%s/%s@%s:%d/%s", user, password, host, port, service)
	connector := go_ora.NewConnector(dsn)
	db := sql.OpenDB(connector)
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("oracle: ping connection: %w", err)
	}
	return &oracleConn{db: db}, nil
}

// oracleConn adapts a database/sql database to the driver connection port.
type oracleConn struct {
	db *sql.DB
}

// Close closes the underlying database connection.
func (c *oracleConn) Close() error {
	return c.db.Close()
}

// Catalog returns the metadata catalog for the connection.
func (c *oracleConn) Catalog() driver.Catalog {
	return &oracleCatalog{db: c.db}
}

// KV returns nil because Oracle is not a key-value store.
func (c *oracleConn) KV() driver.KV {
	return nil
}

// QueryContext executes a query that returns rows.
func (c *oracleConn) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return c.db.QueryContext(ctx, query, args...)
}

// ExecContext executes a statement that does not return rows.
func (c *oracleConn) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return c.db.ExecContext(ctx, query, args...)
}

// oracleCatalog implements driver.Catalog for Oracle.
type oracleCatalog struct {
	db *sql.DB
}

// Databases lists databases visible to the connection.
func (c *oracleCatalog) Databases(ctx context.Context) ([]string, error) {
	rows, err := c.db.QueryContext(ctx, "SELECT db_name FROM v$database")
	if err != nil {
		return nil, fmt.Errorf("oracle: list databases: %w", err)
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("oracle: scan database: %w", err)
		}
		out = append(out, name)
	}
	return out, rows.Err()
}

// Tables lists tables and views for a schema.
func (c *oracleCatalog) Tables(ctx context.Context, schema string) ([]models.TableInfo, error) {
	rows, err := c.db.QueryContext(ctx, "SELECT table_name, 'table' FROM user_tables")
	if err != nil {
		return nil, fmt.Errorf("oracle: list tables: %w", err)
	}
	defer rows.Close()
	var out []models.TableInfo
	for rows.Next() {
		var name, tableType string
		if err := rows.Scan(&name, &tableType); err != nil {
			return nil, fmt.Errorf("oracle: scan table: %w", err)
		}
		out = append(out, models.TableInfo{
			Schema: schema,
			Name:   name,
			Type:   tableType,
		})
	}
	return out, rows.Err()
}

// Columns lists columns for a table.
func (c *oracleCatalog) Columns(ctx context.Context, schema string, table string) ([]models.ColumnInfo, error) {
	rows, err := c.db.QueryContext(ctx, "SELECT column_name, data_type FROM user_tab_columns WHERE table_name = ? ORDER BY column_id", table)
	if err != nil {
		return nil, fmt.Errorf("oracle: list columns: %w", err)
	}
	defer rows.Close()
	var out []models.ColumnInfo
	for rows.Next() {
		var name, dataType string
		if err := rows.Scan(&name, &dataType); err != nil {
			return nil, fmt.Errorf("oracle: scan column: %w", err)
		}
		out = append(out, models.ColumnInfo{
			Name: name,
			Type: dataType,
		})
	}
	return out, rows.Err()
}

// ForeignKeys lists foreign keys for a table.
func (c *oracleCatalog) ForeignKeys(ctx context.Context, schema string, table string) ([]models.ForeignKeyInfo, error) {
	rows, err := c.db.QueryContext(ctx, "SELECT kc.constraint_name, kccu.table_name, kccu.column_name FROM user_constraints kc JOIN user_cons_columns kccu ON kc.constraint_name = kccu.constraint_name WHERE kc.owner = ? AND kc.constraint_type = 'FOREIGN KEY'", schema)
	if err != nil {
		return nil, fmt.Errorf("oracle: list foreign keys: %w", err)
	}
	defer rows.Close()
	var out []models.ForeignKeyInfo
	for rows.Next() {
		var name, refTable, refColumn string
		if err := rows.Scan(&name, &refTable, &refColumn); err != nil {
			return nil, fmt.Errorf("oracle: scan foreign key: %w", err)
		}
		out = append(out, models.ForeignKeyInfo{
			Name:       name,
			Columns:    []string{refColumn},
			RefTable:   refTable,
			RefColumns: []string{refColumn},
		})
	}
	return out, rows.Err()
}

// Indexes lists indexes for a table.
func (c *oracleCatalog) Indexes(ctx context.Context, schema string, table string) ([]string, error) {
	rows, err := c.db.QueryContext(ctx, "SELECT index_name FROM user_indexes WHERE table_name = ? ORDER BY index_name", table)
	if err != nil {
		return nil, fmt.Errorf("oracle: list indexes: %w", err)
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("oracle: scan index: %w", err)
		}
		out = append(out, name)
	}
	return out, rows.Err()
}
