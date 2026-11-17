// Package config defines the connection profile stores and secret store ports.
package config

import (
	"github.com/collins/GoDBClient/internal/models"
)

// SecretStore provides credential persistence via the OS keychain.
type SecretStore interface {
	Get(profileID string) (string, error)
	Set(profileID string, secret string) error
	Delete(profileID string) error
}

// ProfileStore persists connection profiles. Secrets are never stored here.
type ProfileStore interface {
	List() ([]*models.Profile, error)
	Save(profile *models.Profile) error
	Delete(id string) error
	Get(id string) (*models.Profile, error)
}
