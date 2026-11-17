// Package mysql implements the MySQL driver for DataPulse.
package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/collins/DataPulse/internal/driver"
	"github.com/collins/DataPulse/internal/models"
	_ "github.com/go-sql-driver/mysql"
	"github.com/zalando/go-keyring"
)

const (
	serviceName = "DataPulse"
	defaultHost = "localhost"
	defaultPort = 3306
	defaultDB   = "mysql"
)

// Driver opens MySQL connections.
type Driver struct{}

// NewDriver returns a MySQL driver.
func NewDriver() *Driver {
	return &Driver{}
}

// Type returns the MySQL driver type.
func (d *Driver) Type() driver.Type {
	return driver.MySQL
}

// Open opens a MySQL connection from a profile.
func (d *Driver) Open(ctx context.Context, profile *models.Profile) (driver.Conn, error) {
	password, err := keyring.Get(serviceName, profile.ID)
	if err != nil {
		return nil, fmt.Errorf("mysql: read keychain secret: %w", err)
	}
	host := profile.Host
	if host == "" {
		host = defaultHost
	}
	port := profile.Port
	if port == 0 {
		port = defaultPort
	}
	database := profile.Database
	if database == "" {
		database = defaultDB
	}
	user := profile.Username
	if user == "" {
		user = "root"
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true", user, password, host, port, database)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("mysql: open connection: %w", err)
	}
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("mysql: ping connection: %w", err)
	}
	return &mysqlConn{db: db}, nil
}

// mysqlConn adapts a database/sql database to the driver connection port.
type mysqlConn struct {
	db *sql.DB
}

// Close closes the underlying database connection.
func (c *mysqlConn) Close() error {
	return c.db.Close()
}

// Catalog returns the metadata catalog for the connection.
func (c *mysqlConn) Catalog() driver.Catalog {
	return &mysqlCatalog{db: c.db}
}

// KV returns nil because MySQL is not a key-value store.
func (c *mysqlConn) KV() driver.KV {
	return nil
}

// QueryContext executes a query that returns rows.
func (c *mysqlConn) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return c.db.QueryContext(ctx, query, args...)
}

// ExecContext executes a statement that does not return rows.
func (c *mysqlConn) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return c.db.ExecContext(ctx, query, args...)
}

// mysqlCatalog implements driver.Catalog for MySQL.
type mysqlCatalog struct {
	db *sql.DB
}

// Databases lists databases visible to the connection.
func (c *mysqlCatalog) Databases(ctx context.Context) ([]string, error) {
	rows, err := c.db.QueryContext(ctx, "SELECT SCHEMA_NAME FROM information_schema.SCHEMATA ORDER BY SCHEMA_NAME")
	if err != nil {
		return nil, fmt.Errorf("mysql: list databases: %w", err)
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("mysql: scan database: %w", err)
		}
		out = append(out, name)
	}
	return out, rows.Err()
}

// Tables lists tables and views for a schema.
func (c *mysqlCatalog) Tables(ctx context.Context, schema string) ([]models.TableInfo, error) {
	rows, err := c.db.QueryContext(ctx, "SELECT table_schema, table_name, table_type FROM information_schema.tables WHERE table_schema = ? ORDER BY table_schema, table_name", schema)
	if err != nil {
		return nil, fmt.Errorf("mysql: list tables: %w", err)
	}
	defer rows.Close()
	var out []models.TableInfo
	for rows.Next() {
		var tableSchema, name, tableType string
		if err := rows.Scan(&tableSchema, &name, &tableType); err != nil {
			return nil, fmt.Errorf("mysql: scan table: %w", err)
		}
		out = append(out, models.TableInfo{
			Schema: tableSchema,
			Name:   name,
			Type:   tableType,
		})
	}
	return out, rows.Err()
}

// Columns lists columns for a table.
func (c *mysqlCatalog) Columns(ctx context.Context, schema string, table string) ([]models.ColumnInfo, error) {
	rows, err := c.db.QueryContext(ctx, "SELECT column_name, data_type, is_nullable, column_default FROM information_schema.columns WHERE table_schema = ? AND table_name = ? ORDER BY ordinal_position", schema, table)
	if err != nil {
		return nil, fmt.Errorf("mysql: list columns: %w", err)
	}
	defer rows.Close()
	var out []models.ColumnInfo
	for rows.Next() {
		var name, dataType string
		var nullable bool
		var defaultValue any
		if err := rows.Scan(&name, &dataType, &nullable, &defaultValue); err != nil {
			return nil, fmt.Errorf("mysql: scan column: %w", err)
		}
		out = append(out, models.ColumnInfo{
			Name:     name,
			Type:     dataType,
			NotNull:  !nullable,
			Nullable: nullable,
			Default:  defaultValue,
		})
	}
	return out, rows.Err()
}

// ForeignKeys lists foreign keys for a table.
func (c *mysqlCatalog) ForeignKeys(ctx context.Context, schema string, table string) ([]models.ForeignKeyInfo, error) {
	rows, err := c.db.QueryContext(ctx, "SELECT kcu.column_name, ccu.schema_name, ccu.table_name, ccu.column_name FROM information_schema.table_constraints tc JOIN information_schema.key_column_usage kcu ON tc.constraint_name = kcu.constraint_name AND tc.table_schema = kcu.table_schema JOIN information_schema.constraint_column_usage ccu ON tc.constraint_name = ccu.constraint_name WHERE tc.table_schema = ? AND tc.table_name = ? AND tc.constraint_type = 'FOREIGN KEY'", schema, table)
	if err != nil {
		return nil, fmt.Errorf("mysql: list foreign keys: %w", err)
	}
	defer rows.Close()
	var out []models.ForeignKeyInfo
	for rows.Next() {
		var column, refSchema, refTable, refColumn string
		if err := rows.Scan(&column, &refSchema, &refTable, &refColumn); err != nil {
			return nil, fmt.Errorf("mysql: scan foreign key: %w", err)
		}
		out = append(out, models.ForeignKeyInfo{
			Name:       column,
			Columns:    []string{column},
			RefTable:   refTable,
			RefSchema:  refSchema,
			RefColumns: []string{refColumn},
		})
	}
	return out, rows.Err()
}

// Indexes lists indexes for a table.
func (c *mysqlCatalog) Indexes(ctx context.Context, schema string, table string) ([]string, error) {
	rows, err := c.db.QueryContext(ctx, "SELECT index_name FROM information_schema.statistics WHERE table_schema = ? AND table_name = ? ORDER BY index_name", schema, table)
	if err != nil {
		return nil, fmt.Errorf("mysql: list indexes: %w", err)
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("mysql: scan index: %w", err)
		}
		out = append(out, name)
	}
	return out, rows.Err()
}
