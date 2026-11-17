// Package query implements query execution and the destructive-statement guard.
package query

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/collins/GoDBClient/internal/connpool"
	"github.com/collins/GoDBClient/internal/driver"
	"github.com/collins/GoDBClient/internal/models"
)

var (
	ErrNotConnected = errors.New("not connected")
	ErrTimeout      = errors.New("query timeout")
	ErrNeedsConfirm = errors.New("destructive statement requires confirmation")
)

// Executor runs queries through a pooled connection.
type Executor struct {
	driver  driver.Driver
	pool    *connpool.Manager
	mu      sync.Mutex
	cancels map[string]context.CancelFunc
}

// NewExecutor binds an executor to a driver and a connection pool.
func NewExecutor(d driver.Driver, p *connpool.Manager) *Executor {
	return &Executor{driver: d, pool: p, cancels: map[string]context.CancelFunc{}}
}

// Execute runs a single statement with guard checks, timeout, and paging.
func (e *Executor) Execute(ctx context.Context, req models.QueryRequest) (*models.Result, error) {
	return e.ExecuteWithID(ctx, req, NewQueryID())
}

// ExecuteWithID runs a single statement with guard checks, timeout, and paging using a caller-provided query ID.
func (e *Executor) ExecuteWithID(ctx context.Context, req models.QueryRequest, queryID string) (*models.Result, error) {
	if strings.TrimSpace(req.SQL) == "" {
		return nil, errors.New("query: sql is required")
	}
	decision, err := Inspect(req.SQL)
	if err != nil {
		return nil, err
	}
	if decision.NeedsConfirm && !req.Confirm {
		return nil, &DestructiveError{Decision: *decision}
	}
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 100
	}
	if req.PageSize > 500 {
		req.PageSize = 500
	}
	conn, err := e.pool.Get(req.ConnectionID)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	sqlExec, ok := conn.(driver.SQLExec)
	if !ok {
		return nil, fmt.Errorf("query: driver %q does not support sql execution", e.driver.Type())
	}
	ctx, cancel := context.WithTimeout(ctx, req.Timeout)
	defer cancel()
	if strings.TrimSpace(queryID) == "" {
		queryID = NewQueryID()
	}
	e.mu.Lock()
	e.cancels[queryID] = cancel
	e.mu.Unlock()
	defer func() {
		e.mu.Lock()
		delete(e.cancels, queryID)
		e.mu.Unlock()
	}()
	start := time.Now()
	action := strings.ToUpper(strings.TrimSpace(decision.Action))
	if action != "SELECT" && action != "WITH" && action != "SHOW" && action != "DESCRIBE" && action != "DESC" && action != "EXPLAIN" && action != "PRAGMA" && action != "VALUES" {
		sqlStmt := strings.TrimSpace(strings.TrimRight(strings.TrimSpace(req.SQL), ";"))
		res, err := sqlExec.ExecContext(ctx, sqlStmt)
		if err != nil {
			return nil, fmt.Errorf("query: %w", err)
		}
		affected, _ := res.RowsAffected()
		return &models.Result{QueryID: queryID, RowCount: affected, DurationMS: time.Since(start).Milliseconds()}, nil
	}
	rows, err := sqlExec.QueryContext(ctx, pagingSQL(e.driver.Type(), req.SQL, req.Page, req.PageSize))
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	result := &models.Result{Columns: columns, QueryID: queryID}
	for rows.Next() {
		values := make([]any, len(columns))
		dest := make([]any, len(columns))
		for i := range values {
			dest[i] = &values[i]
		}
		if err := rows.Scan(dest...); err != nil {
			return nil, fmt.Errorf("query: %w", err)
		}
		result.Rows = append(result.Rows, values)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	result.RowCount = int64(len(result.Rows))
	result.Truncated = int64(len(result.Rows)) >= int64(req.PageSize) && req.PageSize > 0
	result.DurationMS = time.Since(start).Milliseconds()
	return result, nil
}

// Cancel stops a running query by its query ID.
func (e *Executor) Cancel(queryID string) {
	e.mu.Lock()
	cancel, ok := e.cancels[queryID]
	e.mu.Unlock()
	if ok {
		cancel()
	}
}

// pagingSQL wraps a statement in a dialect-aware page clause.
func pagingSQL(d driver.Type, sql string, page int, pageSize int) string {
	sql = strings.TrimSpace(sql)
	sql = strings.TrimRight(sql, ";")
	sql = strings.TrimSpace(sql)
	offset := int64((page - 1) * pageSize)
	if d == driver.Oracle {
		return fmt.Sprintf("SELECT * FROM (%s) OFFSET %d ROWS FETCH NEXT %d ROWS ONLY", sql, offset, pageSize)
	}
	return fmt.Sprintf("SELECT * FROM (%s) AS _sub LIMIT %d OFFSET %d", sql, int64(pageSize), offset)
}

// NewQueryID returns a random query identifier.
func NewQueryID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b[:])
}
