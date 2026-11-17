package query

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/collins/DataPulse/internal/connpool"
	"github.com/collins/DataPulse/internal/driver"
	"github.com/collins/DataPulse/internal/driver/sqlite"
	"github.com/collins/DataPulse/internal/models"
)

func TestExecutorDestructiveGuardAndExecute(t *testing.T) {
	ctx := context.Background()
	d := sqlite.NewDriver()
	profile := &models.Profile{ID: "test-sqlite", Driver: models.DriverSQLite, Database: ":memory:"}
	conn, err := d.Open(ctx, profile)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer conn.Close()

	// Setup via direct SQLExec (executor wraps queries for paging, which breaks DDL/INSERT).
	sqlExec, ok := conn.(driver.SQLExec)
	if !ok {
		t.Fatal("sqlite conn does not implement SQLExec")
	}
	if _, err := sqlExec.ExecContext(ctx, "CREATE TABLE t(id INTEGER)"); err != nil {
		t.Fatalf("create table: %v", err)
	}
	if _, err := sqlExec.ExecContext(ctx, "INSERT INTO t(id) VALUES (1)"); err != nil {
		t.Fatalf("insert: %v", err)
	}

	pool := connpool.NewManager()
	if err := pool.Connect(ctx, profile.ID, conn); err != nil {
		t.Fatalf("pool connect: %v", err)
	}
	ex := NewExecutor(d, pool)

	// destructive without confirm must be blocked before execution
	_, err = ex.ExecuteWithID(ctx, models.QueryRequest{ConnectionID: profile.ID, SQL: "DELETE FROM t", Page: 1, PageSize: 100, Timeout: 5 * time.Second}, "del1")
	var derr *DestructiveError
	if !errors.As(err, &derr) {
		t.Fatalf("expected DestructiveError, got %v", err)
	}
	if derr.Decision.Action != "DELETE" || !derr.Decision.NeedsConfirm {
		t.Fatalf("unexpected decision: %+v", derr.Decision)
	}

	// confirmed destructive executes
	res, err := ex.ExecuteWithID(ctx, models.QueryRequest{ConnectionID: profile.ID, SQL: "DELETE FROM t", Page: 1, PageSize: 100, Timeout: 5 * time.Second, Confirm: true}, "del2")
	if err != nil {
		t.Fatalf("confirmed delete: %v", err)
	}
	if res == nil || res.QueryID != "del2" {
		t.Fatalf("unexpected result: %+v", res)
	}

	// normal select executes through shared executor path
	res, err = ex.ExecuteWithID(ctx, models.QueryRequest{ConnectionID: profile.ID, SQL: "SELECT id FROM t", Page: 1, PageSize: 100, Timeout: 5 * time.Second}, "sel1")
	if err != nil {
		t.Fatalf("select: %v", err)
	}
	if res == nil || res.QueryID != "sel1" {
		t.Fatalf("unexpected result: %+v", res)
	}
	if len(res.Rows) != 0 {
		t.Fatalf("expected 0 rows after delete, got %d", len(res.Rows))
	}
}
