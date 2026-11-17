// Package models defines the shared data transfer objects exchanged between
// the backend core and the frontend bridge layers.
package models

import "time"

// DriverType identifies the family of database a connection targets.
type DriverType string

const (
	DriverPostgres DriverType = "postgres"
	DriverMySQL    DriverType = "mysql"
	DriverSQLite   DriverType = "sqlite"
	DriverOracle   DriverType = "oracle"
	DriverRedis    DriverType = "redis"
)

// Profile describes a database connection. Credentials are intentionally not
// part of the persisted profile; they live in the OS keychain.
type Profile struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Driver   DriverType        `json:"driver"`
	Host     string            `json:"host,omitempty"`
	Port     int               `json:"port,omitempty"`
	Database string            `json:"database,omitempty"`
	Username string            `json:"username,omitempty"`
	SSLMode  string            `json:"sslmode,omitempty"`
	Options  map[string]string `json:"options,omitempty"`
}

// ColumnInfo describes a single column of a table or view.
type ColumnInfo struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	NotNull   bool   `json:"notnull"`
	Nullable  bool   `json:"nullable"`
	Default   any    `json:"default,omitempty"`
	IsPrimary bool   `json:"isPrimary"`
}

// ForeignKeyInfo describes a foreign key relationship.
type ForeignKeyInfo struct {
	Name       string   `json:"name"`
	Columns    []string `json:"columns"`
	RefTable   string   `json:"referencedTable"`
	RefSchema  string   `json:"referencedSchema,omitempty"`
	RefColumns []string `json:"referencedColumns"`
}

// TableInfo describes a table or view for the metadata explorer.
type TableInfo struct {
	Schema      string           `json:"schema"`
	Name        string           `json:"name"`
	Type        string           `json:"type"` // "table" or "view"
	Columns     []ColumnInfo     `json:"columns"`
	PrimaryKey  []string         `json:"primaryKey"`
	ForeignKeys []ForeignKeyInfo `json:"foreignKeys"`
	Indexes     []string         `json:"indexes"`
}

// KeyInfo describes a Redis key: type, TTL, and value length.
type KeyInfo struct {
	Key    string `json:"key"`
	Type   string `json:"type"`
	TTL    int64  `json:"ttlSeconds,omitempty"`
	Length int64  `json:"length,omitempty"`
}

// ValueView is a JSON-safe view of a stored Redis value.
type ValueView struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Kind  string `json:"kind"`
}

// Result is the structured, JSON-safe representation of a query result.
type Result struct {
	Columns    []string `json:"columns"`
	Rows       [][]any  `json:"rows"`
	RowCount   int64    `json:"rowCount"`
	Truncated  bool     `json:"truncated"`
	DurationMS int64    `json:"durationMs"`
	QueryID    string   `json:"queryId"`
}

// GuardDecision describes the safety verdict for a SQL statement.
type GuardDecision struct {
	Action       string `json:"action"`
	NeedsConfirm bool   `json:"needsConfirm"`
}

// QueryRequest describes a single query execution.
type QueryRequest struct {
	ConnectionID string
	SQL          string
	Page         int
	PageSize     int
	Timeout      time.Duration
	Confirm      bool
}
