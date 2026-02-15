package main

import (
	"os"
	"path/filepath"
	"time"
)

const maxBodySize = 1 << 20 // 1MB

var (
	port            = envOr("BRIDGE_PORT", "18800")
	cdpURL          = os.Getenv("CDP_URL") // empty = launch Chrome ourselves
	token           = os.Getenv("BRIDGE_TOKEN")
	stateDir        = envOr("BRIDGE_STATE_DIR", filepath.Join(homeDir(), ".browser-bridge"))
	headless        = os.Getenv("BRIDGE_HEADLESS") == "true"
	noRestore       = os.Getenv("BRIDGE_NO_RESTORE") == "true"
	profileDir      = envOr("BRIDGE_PROFILE", filepath.Join(homeDir(), ".browser-bridge", "chrome-profile"))
	actionTimeout   = 15 * time.Second
	shutdownTimeout = 10 * time.Second
)

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func homeDir() string {
	h, _ := os.UserHomeDir()
	return h
}
