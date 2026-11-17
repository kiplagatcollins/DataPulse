// Package bridge adapts the backend core services to application-facing APIs.
package bridge

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/collins/DataPulse/internal/config"
	"github.com/collins/DataPulse/internal/connpool"
	"github.com/collins/DataPulse/internal/driver"
	"github.com/collins/DataPulse/internal/metadata"
	"github.com/collins/DataPulse/internal/models"
	"github.com/collins/DataPulse/internal/query"
)

// ConnectionService coordinates profile and connection lifecycle operations.
type ConnectionService struct {
	pool     *connpool.Manager
	registry *driver.Registry
	profiles config.ProfileStore
	secrets  config.SecretStore
}

// NewConnectionService returns a connection service.
func NewConnectionService(p *connpool.Manager, r *driver.Registry, ps config.ProfileStore, ss config.SecretStore) *ConnectionService {
	return &ConnectionService{pool: p, registry: r, profiles: ps, secrets: ss}
}

// ListProfiles returns all saved connection profiles.
func (s *ConnectionService) ListProfiles() ([]*models.Profile, error) {
	return s.profiles.List()
}

// SaveProfile persists a connection profile and its keychain secret.
func (s *ConnectionService) SaveProfile(ctx context.Context, profile *models.Profile, password string) error {
	if profile.ID == "" {
		profile.ID = newID()
	}
	if password != "" || profile.Driver == models.DriverSQLite {
		if err := s.secrets.Set(profile.ID, password); err != nil {
			return fmt.Errorf("save connection %q: %w", profile.ID, err)
		}
	}
	if err := s.profiles.Save(profile); err != nil {
		return fmt.Errorf("save connection %q: %w", profile.ID, err)
	}
	return nil
}

// DeleteProfile removes a connection profile and its keychain secret.
func (s *ConnectionService) DeleteProfile(ctx context.Context, id string) error {
	if err := s.secrets.Delete(id); err != nil {
		return fmt.Errorf("delete connection %q: %w", id, err)
	}
	if err := s.profiles.Delete(id); err != nil {
		return fmt.Errorf("delete connection %q: %w", id, err)
	}
	return nil
}

// TestConnection verifies connectivity with a bounded timeout.
func (s *ConnectionService) TestConnection(ctx context.Context, profile *models.Profile, password string) error {
	if profile == nil {
		return fmt.Errorf("test connection: nil profile")
	}
	d, err := s.registry.Get(driver.Type(profile.Driver))
	if err != nil {
		return fmt.Errorf("test connection: %w", err)
	}

	testProfile := *profile
	tempSecret := false
	if profile.Driver != models.DriverSQLite {
		if testProfile.ID == "" {
			testProfile.ID = "test-" + newID()
		}
		if err := s.secrets.Set(testProfile.ID, password); err != nil {
			return fmt.Errorf("test connection: store temporary secret: %w", err)
		}
		tempSecret = true
		defer func() { _ = s.secrets.Delete(testProfile.ID) }()
	}

	conn, err := d.Open(ctx, &testProfile)
	if err != nil {
		return fmt.Errorf("test connection: %w", err)
	}
	if err := conn.Close(); err != nil {
		return fmt.Errorf("test connection: %w", err)
	}
	_ = tempSecret
	return nil
}

// Connect opens a connection for a saved profile and registers it in the pool.
func (s *ConnectionService) Connect(ctx context.Context, id string) error {
	profile, err := s.profiles.Get(id)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	d, err := s.registry.Get(driver.Type(profile.Driver))
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	conn, err := d.Open(ctx, profile)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	if err := s.pool.Connect(ctx, id, conn); err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	return nil
}

// Disconnect closes the live connection for a profile.
func (s *ConnectionService) Disconnect(ctx context.Context, id string) error {
	return s.pool.Disconnect(id)
}

// ListActive returns the IDs of live connections.
func (s *ConnectionService) ListActive() ([]string, error) {
	return s.pool.List(), nil
}

// MetadataService resolves metadata for a live connection.
type MetadataService struct {
	Manager *metadata.Manager
}

// NewMetadataService returns a metadata service.
func NewMetadataService(s *metadata.Manager) *MetadataService {
	return &MetadataService{Manager: s}
}

// Databases lists the databases or schemas visible to the connection.
func (s *MetadataService) Databases(ctx context.Context, connID string) ([]string, error) {
	return s.Manager.Databases(ctx, connID)
}

// Tables lists tables and views for a schema.
func (s *MetadataService) Tables(ctx context.Context, connID string, schema string) ([]models.TableInfo, error) {
	return s.Manager.Tables(ctx, connID, schema)
}

// Columns lists columns for a table.
func (s *MetadataService) Columns(ctx context.Context, connID string, schema string, table string) ([]models.ColumnInfo, error) {
	return s.Manager.Columns(ctx, connID, schema, table)
}

// ForeignKeys lists foreign keys for a table.
func (s *MetadataService) ForeignKeys(ctx context.Context, connID string, schema string, table string) ([]models.ForeignKeyInfo, error) {
	return s.Manager.ForeignKeys(ctx, connID, schema, table)
}

// Indexes lists indexes for a table.
func (s *MetadataService) Indexes(ctx context.Context, connID string, schema string, table string) ([]string, error) {
	return s.Manager.Indexes(ctx, connID, schema, table)
}

type activeQuery struct {
	queryID  string
	executor *query.Executor
}

// QueryService executes queries against a live connection.
type QueryService struct {
	pool      *connpool.Manager
	registry  *driver.Registry
	profiles  config.ProfileStore
	mu        sync.Mutex
	executors map[driver.Type]*query.Executor
	active    map[string]activeQuery
}

// NewQueryService returns a query service.
func NewQueryService(p *connpool.Manager, r *driver.Registry, ps config.ProfileStore) *QueryService {
	return &QueryService{
		pool:      p,
		registry:  r,
		profiles:  ps,
		executors: map[driver.Type]*query.Executor{},
		active:    map[string]activeQuery{},
	}
}

func (s *QueryService) executorFor(dt driver.Type) (*query.Executor, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if ex, ok := s.executors[dt]; ok {
		return ex, nil
	}
	d, err := s.registry.Get(dt)
	if err != nil {
		return nil, err
	}
	ex := query.NewExecutor(d, s.pool)
	s.executors[dt] = ex
	return ex, nil
}

// Execute runs a query for the connection referenced by the request.
func (s *QueryService) Execute(ctx context.Context, req models.QueryRequest) (*models.Result, error) {
	profile, err := s.profiles.Get(req.ConnectionID)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	executor, err := s.executorFor(driver.Type(profile.Driver))
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	queryID := query.NewQueryID()
	s.mu.Lock()
	s.active[req.ConnectionID] = activeQuery{queryID: queryID, executor: executor}
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		if cur, ok := s.active[req.ConnectionID]; ok && cur.queryID == queryID {
			delete(s.active, req.ConnectionID)
		}
		s.mu.Unlock()
	}()
	return executor.ExecuteWithID(ctx, req, queryID)
}

// Inspect classifies a SQL statement for destructive-guard confirmation.
func (s *QueryService) Inspect(ctx context.Context, sql string) (*models.GuardDecision, error) {
	return query.Inspect(sql)
}

// Cancel stops the active query for a connection.
func (s *QueryService) Cancel(ctx context.Context, connID string) error {
	s.mu.Lock()
	cur, ok := s.active[connID]
	s.mu.Unlock()
	if ok {
		cur.executor.Cancel(cur.queryID)
	}
	return nil
}

// RedisService browses and executes against key-value connections.
type RedisService struct {
	Pool *connpool.Manager
}

// NewRedisService returns a redis service.
func NewRedisService(p *connpool.Manager) *RedisService {
	return &RedisService{Pool: p}
}

// ScanKeys lists keys matching a pattern using cursor-based scan.
func (s *RedisService) ScanKeys(ctx context.Context, connID string, pattern string, cursor uint64, count int64) ([]string, uint64, error) {
	conn, err := s.Pool.Get(connID)
	if err != nil {
		return nil, 0, fmt.Errorf("redis: %w", err)
	}
	kv := conn.KV()
	if kv == nil {
		return nil, 0, fmt.Errorf("redis: connection %q does not support key-value access", connID)
	}
	return kv.ScanKeys(ctx, pattern, cursor, count)
}

// KeyInfo returns metadata for a single key.
func (s *RedisService) KeyInfo(ctx context.Context, connID string, key string) (models.KeyInfo, error) {
	conn, err := s.Pool.Get(connID)
	if err != nil {
		return models.KeyInfo{}, fmt.Errorf("redis: %w", err)
	}
	kv := conn.KV()
	if kv == nil {
		return models.KeyInfo{}, fmt.Errorf("redis: connection %q does not support key-value access", connID)
	}
	return kv.KeyInfo(ctx, key)
}

// GetValue returns a single key value.
func (s *RedisService) GetValue(ctx context.Context, connID string, key string) (models.ValueView, error) {
	conn, err := s.Pool.Get(connID)
	if err != nil {
		return models.ValueView{}, fmt.Errorf("redis: %w", err)
	}
	kv := conn.KV()
	if kv == nil {
		return models.ValueView{}, fmt.Errorf("redis: connection %q does not support key-value access", connID)
	}
	return kv.GetValue(ctx, key)
}

// Exec runs a raw key-value command.
func (s *RedisService) Exec(ctx context.Context, connID string, args []string) ([]string, error) {
	conn, err := s.Pool.Get(connID)
	if err != nil {
		return nil, fmt.Errorf("redis: %w", err)
	}
	kv := conn.KV()
	if kv == nil {
		return nil, fmt.Errorf("redis: connection %q does not support key-value access", connID)
	}
	return kv.Exec(ctx, args)
}

// newID returns a new random identifier.
func newID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return hex.EncodeToString(b[:])
	}
	return hex.EncodeToString(b[:])
}
