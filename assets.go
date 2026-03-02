package main

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func (cfg apiConfig) ensureAssetsDir() error {
	if _, err := os.Stat(cfg.assetsRoot); os.IsNotExist(err) {
		return os.Mkdir(cfg.assetsRoot, 0o755)
	}
	return nil
}

func generateAssetName() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return base64.RawURLEncoding.EncodeToString(bytes)
}

func (cfg apiConfig) getAssetPath(ext string) string {
	return fmt.Sprintf("%s%s", filepath.Join(cfg.assetsRoot, generateAssetName()), ext)
}

func (cfg apiConfig) getAssetURL(path string) string {
	return fmt.Sprintf("http://localhost:%s/%s", cfg.port, path)
}

func (cfg apiConfig) getS3URL(key string) string {
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", cfg.s3Bucket, cfg.s3Region, key)
}

func getVideoAspectRatio(filepath string) (string, error) {
	type Stream struct {
		Height             int    `json:"height"`
		Width              int    `json:"width"`
		DisplayAspectRatio string `json:"display_aspect_ratio"`
	}

	type ProbeResult struct {
		Streams []Stream `json:"streams"`
	}

	cmd := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filepath)

	var stdoutBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf

	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	err := cmd.Run()
	if err != nil {
		return "other", fmt.Errorf("error: %s\nstderr: %s", err, stderrBuf.String())
	}

	var result ProbeResult
	err = json.Unmarshal(stdoutBuf.Bytes(), &result)
	if err != nil || len(result.Streams) < 1 {
		return "other", fmt.Errorf("error ffprobing: %s", err)
	}

	return func() string {
		switch result.Streams[0].DisplayAspectRatio {
		case "16:9":
			return "landscape"
		case "9:16":
			return "portrait"
		default:
			return "other"
		}
	}(), nil
}
