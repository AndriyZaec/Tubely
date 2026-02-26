package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

func (cfg apiConfig) ensureAssetsDir() error {
	if _, err := os.Stat(cfg.assetsRoot); os.IsNotExist(err) {
		return os.Mkdir(cfg.assetsRoot, 0o755)
	}
	return nil
}

func (cfg apiConfig) getAssetPath(assetID uuid.UUID, ext string) string {
	return fmt.Sprintf("%s%s", filepath.Join(cfg.assetsRoot, assetID.String()), ext)
}

func (cfg apiConfig) getAssetURL(path string) string {
	return fmt.Sprintf("http://localhost:%s/%s", cfg.port, path)
}
