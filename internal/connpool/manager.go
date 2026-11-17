// Package connpool manages live database connections keyed by profile ID.
package connpool

import (
	"context"
	"errors"
	"sync"

	"github.com/collins/DataPulse/internal/driver"
)

var ErrNotConnected = errors.New("not connected")

// Manager tracks the live connection for each profile ID.
type Manager struct {
	mu    sync.RWMutex
	conns map[string]driver.Conn
}

// NewManager returns a manager that starts with no live connections.
func NewManager() *Manager {
	return &Manager{conns: map[string]driver.Conn{}}
}

// Connect registers conn as the live connection for id.
func (m *Manager) Connect(ctx context.Context, id string, conn driver.Conn) error {
	if conn == nil {
		return errors.New("connpool: nil connection")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.conns[id] = conn
	return nil
}

// Disconnect removes the live connection for id.
func (m *Manager) Disconnect(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.conns, id)
	return nil
}

// Get returns the live connection for id.
func (m *Manager) Get(id string) (driver.Conn, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	conn, ok := m.conns[id]
	if !ok {
		return nil, ErrNotConnected
	}
	return conn, nil
}

// CloseAll drops all live connections.
func (m *Manager) CloseAll() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.conns = map[string]driver.Conn{}
	return nil
}

// List returns the IDs of live connections.
func (m *Manager) List() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	ids := make([]string, 0, len(m.conns))
	for id := range m.conns {
		ids = append(ids, id)
	}
	return ids
}
