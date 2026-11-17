// Package postgres implements the PostgreSQL driver for DataPulse.
package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/collins/DataPulse/internal/driver"
	"github.com/collins/DataPulse/internal/models"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/zalando/go-keyring"
)

const (
	serviceName     = "DataPulse"
	defaultHost     = "localhost"
	defaultPort     = 5432
	defaultDatabase = "postgres"
	defaultSSLMode  = "prefer"
)

// Driver opens PostgreSQL connections.
type Driver struct{}

// NewDriver returns a PostgreSQL driver.
func NewDriver() *Driver {
	return &Driver{}
}

// Type returns the PostgreSQL driver type.
func (d *Driver) Type() driver.Type {
	return driver.Postgres
}

// Open opens a PostgreSQL connection from a profile.
func (d *Driver) Open(ctx context.Context, profile *models.Profile) (driver.Conn, error) {
	password, err := keyring.Get(serviceName, profile.ID)
	if err != nil {
		return nil, fmt.Errorf("postgres: read keychain secret: %w", err)
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
		database = defaultDatabase
	}
	user := profile.Username
	if user == "" {
		user = "postgres"
	}
	sslmode := profile.SSLMode
	if sslmode == "" {
		sslmode = defaultSSLMode
	}
	dsn := fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=%s", host, port, database, user, password, sslmode)
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("postgres: open connection: %w", err)
	}
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("postgres: ping connection: %w", err)
	}
	return &postgresConn{db: db}, nil
}

// postgresConn adapts a database/sql database to the driver connection port.
type postgresConn struct {
	db *sql.DB
}

// Close closes the underlying database connection.
func (c *postgresConn) Close() error {
	return c.db.Close()
}

// Catalog returns the metadata catalog for the connection.
func (c *postgresConn) Catalog() driver.Catalog {
	return &postgresCatalog{db: c.db}
}

// KV returns nil because PostgreSQL is not a key-value store.
func (c *postgresConn) KV() driver.KV {
	return nil
}

// QueryContext executes a query that returns rows.
func (c *postgresConn) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return c.db.QueryContext(ctx, query, args...)
}

// ExecContext executes a statement that does not return rows.
func (c *postgresConn) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return c.db.ExecContext(ctx, query, args...)
}

// postgresCatalog implements driver.Catalog for PostgreSQL.
type postgresCatalog struct {
	db *sql.DB
}

// Databases lists schemas visible to the connection.
func (c *postgresCatalog) Databases(ctx context.Context) ([]string, error) {
	rows, err := c.db.QueryContext(ctx, "SELECT schema_name FROM information_schema.schemata WHERE schema_name NOT LIKE 'pg_%' AND schema_name <> 'information_schema' ORDER BY schema_name")
	if err != nil {
		return nil, fmt.Errorf("postgres: list schemas: %w", err)
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("postgres: scan schema: %w", err)
		}
		out = append(out, name)
	}
	return out, rows.Err()
}

// Tables lists tables and views for a schema.
func (c *postgresCatalog) Tables(ctx context.Context, schema string) ([]models.TableInfo, error) {
	rows, err := c.db.QueryContext(ctx, "SELECT table_schema, table_name, table_type FROM information_schema.tables WHERE table_schema = $1 ORDER BY table_schema, table_name", schema)
	if err != nil {
		return nil, fmt.Errorf("postgres: list tables: %w", err)
	}
	defer rows.Close()
	var out []models.TableInfo
	for rows.Next() {
		var schema, name, tableType string
		if err := rows.Scan(&schema, &name, &tableType); err != nil {
			return nil, fmt.Errorf("postgres: scan table: %w", err)
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
func (c *postgresCatalog) Columns(ctx context.Context, schema string, table string) ([]models.ColumnInfo, error) {
	rows, err := c.db.QueryContext(ctx, "SELECT column_name, data_type, is_nullable, column_default FROM information_schema.columns WHERE table_schema = $1 AND table_name = $2 ORDER BY ordinal_position", schema, table)
	if err != nil {
		return nil, fmt.Errorf("postgres: list columns: %w", err)
	}
	defer rows.Close()
	var out []models.ColumnInfo
	for rows.Next() {
		var name, dataType string
		var nullable bool
		var defaultValue any
		if err := rows.Scan(&name, &dataType, &nullable, &defaultValue); err != nil {
			return nil, fmt.Errorf("postgres: scan column: %w", err)
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
func (c *postgresCatalog) ForeignKeys(ctx context.Context, schema string, table string) ([]models.ForeignKeyInfo, error) {
	rows, err := c.db.QueryContext(ctx, "SELECT kcu.column_name, ccu.table_schema, ccu.table_name, ccu.column_name FROM information_schema.table_constraints tc JOIN information_schema.key_column_usage kcu ON tc.constraint_name = kcu.constraint_name AND tc.table_schema = kcu.table_schema JOIN information_schema.constraint_column_usage ccu ON tc.constraint_name = ccu.constraint_name WHERE tc.table_schema = $1 AND tc.table_name = $2 AND tc.constraint_type = 'FOREIGN KEY'", schema, table)
	if err != nil {
		return nil, fmt.Errorf("postgres: list foreign keys: %w", err)
	}
	defer rows.Close()
	var out []models.ForeignKeyInfo
	for rows.Next() {
		var column, refSchema, refTable, refColumn string
		if err := rows.Scan(&column, &refSchema, &refTable, &refColumn); err != nil {
			return nil, fmt.Errorf("postgres: scan foreign key: %w", err)
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
func (c *postgresCatalog) Indexes(ctx context.Context, schema string, table string) ([]string, error) {
	rows, err := c.db.QueryContext(ctx, "SELECT indexname FROM pg_indexes WHERE schemaname = $1 AND tablename = $2 ORDER BY indexname", schema, table)
	if err != nil {
		return nil, fmt.Errorf("postgres: list indexes: %w", err)
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("postgres: scan index: %w", err)
		}
		out = append(out, name)
	}
	return out, rows.Err()
}
