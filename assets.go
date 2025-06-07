package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func (cfg apiConfig) ensureAssetsDir() error {
	if _, err := os.Stat(cfg.assetsRoot); os.IsNotExist(err) {
		return os.Mkdir(cfg.assetsRoot, 0755)
	}
	return nil
}

func getAssetPath(videoID, mediaType string) string {
	ext := mediaTypeToExt(mediaType)
	return fmt.Sprintf("%s%s", videoID, ext)
}

func (cfg apiConfig) getAssetDiskPath(assetPath string) string {
	return filepath.Join(cfg.assetsRoot, assetPath)
}

func (cfg apiConfig) getAssetURL(assetPath string) string {
	return fmt.Sprintf("http://localhost:%s/assets/%s", cfg.port, assetPath)
}

func mediaTypeToExt(mediaType string) string {
	parts := strings.Split(mediaType, "/")
	if len(parts) != 2 {
		return ".bin"
	}
	return "." + parts[1]
}

func getVideoAspectRation(filePath string) (string, error) {
	cmd := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filePath)
	cmd.Stdout = &bytes.Buffer{}
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("failed to get video aspect ratio, %w", err)
	}
	type Stream struct {
		Width  int `json:"width"`
		Height int `json:"height"`
	}
	type VideoInfo struct {
		Streams []Stream `json:"streams"`
	}
	var videoInfo VideoInfo
	err = json.Unmarshal(cmd.Stdout.(*bytes.Buffer).Bytes(), &videoInfo)
	if err != nil {
		return "", fmt.Errorf("failed to parse ffprobe output, %w", err)
	}
	width := float64(videoInfo.Streams[0].Width)
	height := float64(videoInfo.Streams[0].Height)
	ratio := width / height
	// Check common aspect ratios with tolerance
	if math.Abs(ratio-16.0/9.0) < 0.1 {
		return "16:9", nil
	}
	if math.Abs(ratio-4.0/3.0) < 0.1 {
		return "4:3", nil
	}
	if math.Abs(ratio-9.0/16.0) < 0.1 {
		return "9:16", nil // Portrait/vertical video
	}
	return "other", nil
}
