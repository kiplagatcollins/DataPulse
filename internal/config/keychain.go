// Package config defines the connection profile stores and secret store ports.
package config

import "github.com/zalando/go-keyring"

const serviceName = "DataPulse"

// keychainSecretStore persists credentials via the OS keychain.
type keychainSecretStore struct{}

// NewKeychainSecretStore returns a secret store backed by the OS keychain.
func NewKeychainSecretStore() SecretStore {
	return &keychainSecretStore{}
}

// Get returns the stored secret for a profile.
func (s *keychainSecretStore) Get(profileID string) (string, error) {
	return keyring.Get(serviceName, profileID)
}

// Set stores the secret for a profile.
func (s *keychainSecretStore) Set(profileID string, secret string) error {
	return keyring.Set(serviceName, profileID, secret)
}

// Delete removes the stored secret for a profile.
func (s *keychainSecretStore) Delete(profileID string) error {
	return keyring.Delete(serviceName, profileID)
}
