package serve

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
)

type ServerConfig struct {
	Port     string
	Host     string
	DocsPath string
}

func NewServerConfig() *ServerConfig {
	return &ServerConfig{
		Port:     ":8080",
		Host:     "localhost",
		DocsPath: "docs",
	}
}

func LoadInventory() (*devicetypes.Inventory, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	dbPath := filepath.Join(home, ".cani", "canidb.json")
	content, err := os.ReadFile(dbPath)
	if err != nil {
		log.Printf("failed to read file %s: %v", dbPath, err)
		return &devicetypes.Inventory{}, nil
	}

	var inventory devicetypes.Inventory
	if err := json.Unmarshal(content, &inventory); err != nil {
		return nil, err
	}

	return &inventory, nil
}
