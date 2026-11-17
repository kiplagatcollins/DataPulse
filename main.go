package main

import (
	"embed"
	"log"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v3/pkg/application"

	"github.com/collins/DataPulse/internal/bridge"
	"github.com/collins/DataPulse/internal/config"
	"github.com/collins/DataPulse/internal/connpool"
	"github.com/collins/DataPulse/internal/driver"
	"github.com/collins/DataPulse/internal/driver/mysql"
	"github.com/collins/DataPulse/internal/driver/oracle"
	"github.com/collins/DataPulse/internal/driver/postgres"
	"github.com/collins/DataPulse/internal/driver/redis"
	"github.com/collins/DataPulse/internal/driver/sqlite"
	"github.com/collins/DataPulse/internal/metadata"
)

//go:embed all:frontend/dist
var assets embed.FS

func userConfigPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return filepath.Join("datapulse", "profiles.json")
	}
	return filepath.Join(dir, "datapulse", "profiles.json")
}

func main() {
	pool := connpool.NewManager()
	registry := driver.NewRegistry()
	registry.Register(postgres.NewDriver())
	registry.Register(redis.NewDriver())
	registry.Register(sqlite.NewDriver())
	registry.Register(mysql.NewDriver())
	registry.Register(oracle.NewDriver())
	profiles, err := config.NewFileProfileStore(userConfigPath())
	if err != nil {
		log.Fatalf("DataPulse: open profile store: %v", err)
	}
	secrets := config.NewKeychainSecretStore()
	meta := metadata.NewManager(pool)

	app := application.New(application.Options{
		Name: "DataPulse",
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Services: []application.Service{
			application.NewService(bridge.NewMetadataService(meta)),
			application.NewService(bridge.NewRedisService(pool)),
			application.NewService(bridge.NewConnectionService(pool, registry, profiles, secrets)),
			application.NewService(bridge.NewQueryService(pool, registry, profiles)),
		},
	})

	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:            "DataPulse",
		Width:            1200,
		Height:           800,
		BackgroundColour: application.NewRGB(15, 17, 23),
		URL:              "/",
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
	})

	if err := app.Run(); err != nil {
		log.Fatalf("DataPulse failed to start: %v", err)
	}
}
