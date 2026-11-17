package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"

	"github.com/collins/DataPulse/internal/models"
)

// fileProfileStore persists profiles to a JSON file.
type fileProfileStore struct {
	mu       sync.RWMutex
	path     string
	profiles []*models.Profile
}

// NewFileProfileStore returns a profile store backed by a JSON file.
func NewFileProfileStore(path string) (*fileProfileStore, error) {
	s := &fileProfileStore{path: path}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

// load reads the persisted profiles from disk.
func (s *fileProfileStore) load() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if len(data) == 0 {
		return nil
	}
	return json.Unmarshal(data, &s.profiles)
}

// persist writes the profiles to disk.
func (s *fileProfileStore) persist() error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s.profiles, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o600)
}

// List returns all saved profiles.
func (s *fileProfileStore) List() ([]*models.Profile, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*models.Profile, len(s.profiles))
	copy(out, s.profiles)
	return out, nil
}

// Save persists a profile, upserting by ID.
func (s *fileProfileStore) Save(profile *models.Profile) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, p := range s.profiles {
		if p.ID == profile.ID {
			s.profiles[i] = profile
			return s.persist()
		}
	}
	s.profiles = append(s.profiles, profile)
	return s.persist()
}

// Delete removes the profile with the given ID.
func (s *fileProfileStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := s.profiles[:0]
	for _, p := range s.profiles {
		if p.ID != id {
			out = append(out, p)
		}
	}
	s.profiles = out
	return s.persist()
}

// Get returns the profile with the given ID.
func (s *fileProfileStore) Get(id string) (*models.Profile, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, p := range s.profiles {
		if p.ID == id {
			return p, nil
		}
	}
	return nil, errors.New("profile not found")
}
