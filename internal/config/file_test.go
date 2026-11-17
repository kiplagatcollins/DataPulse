package config

import (
	"path/filepath"
	"testing"

	"github.com/collins/GoDBClient/internal/models"
)

func TestFileProfileStoreSaveGetListDelete(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "profiles.json")
	store, err := NewFileProfileStore(path)
	if err != nil {
		t.Fatalf("NewFileProfileStore: %v", err)
	}

	profile := &models.Profile{
		ID:       "p1",
		Name:     "Local",
		Driver:   models.DriverPostgres,
		Host:     "localhost",
		Port:     5432,
		Database: "app",
	}
	if err := store.Save(profile); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := store.Get("p1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.ID != "p1" || got.Name != "Local" {
		t.Fatalf("unexpected profile: %+v", got)
	}

	list, err := store.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 1 || list[0].ID != "p1" {
		t.Fatalf("unexpected list: %+v", list)
	}

	if err := store.Delete("p1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := store.Get("p1"); err == nil {
		t.Fatal("expected error deleting then getting unknown profile")
	}
}
