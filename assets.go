package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
)

func (cfg apiConfig) ensureAssetsDir() error {
	if _, err := os.Stat(cfg.assetsRoot); os.IsNotExist(err) {
		return os.Mkdir(cfg.assetsRoot, 0o755)
	}
	return nil
}

func (cfg apiConfig) getAssetPath(ext string) string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	randFileName := base64.RawURLEncoding.EncodeToString(bytes)
	return fmt.Sprintf("%s%s", filepath.Join(cfg.assetsRoot, randFileName), ext)
}

func (cfg apiConfig) getAssetURL(path string) string {
	return fmt.Sprintf("http://localhost:%s/%s", cfg.port, path)
}
