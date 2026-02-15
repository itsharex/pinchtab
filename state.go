package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

// TabState represents a saved tab for session persistence.
type TabState struct {
	ID    string `json:"id"`
	URL   string `json:"url"`
	Title string `json:"title"`
}

// SessionState is the on-disk format for saved sessions.
type SessionState struct {
	Tabs    []TabState `json:"tabs"`
	SavedAt string     `json:"savedAt"`
}

// markCleanExit patches Chrome's preferences to prevent "didn't shut down correctly" bar.
func markCleanExit() {
	prefsPath := filepath.Join(profileDir, "Default", "Preferences")
	data, err := os.ReadFile(prefsPath)
	if err != nil {
		return
	}
	patched := strings.ReplaceAll(string(data), `"exit_type":"Crashed"`, `"exit_type":"Normal"`)
	patched = strings.ReplaceAll(patched, `"exited_cleanly":false`, `"exited_cleanly":true`)
	if patched != string(data) {
		if err := os.WriteFile(prefsPath, []byte(patched), 0644); err != nil {
			log.Printf("Failed to patch prefs: %v", err)
		}
	}
}

func saveState() {
	targets, err := listTargets()
	if err != nil {
		log.Printf("Failed to save state: %v", err)
		return
	}

	tabs := make([]TabState, 0, len(targets))
	for _, t := range targets {
		if t.URL != "" && t.URL != "about:blank" && t.URL != "chrome://newtab/" {
			tabs = append(tabs, TabState{
				ID:    string(t.TargetID),
				URL:   t.URL,
				Title: t.Title,
			})
		}
	}

	state := SessionState{
		Tabs:    tabs,
		SavedAt: time.Now().UTC().Format(time.RFC3339),
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		log.Printf("Failed to marshal state: %v", err)
		return
	}
	path := filepath.Join(stateDir, "sessions.json")
	if err := os.WriteFile(path, data, 0644); err != nil {
		log.Printf("Failed to write state: %v", err)
	} else {
		log.Printf("Saved %d tabs to %s", len(tabs), path)
	}
}

func restoreState() {
	path := filepath.Join(stateDir, "sessions.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var state SessionState
	if err := json.Unmarshal(data, &state); err != nil {
		return
	}

	restored := 0
	for _, tab := range state.Tabs {
		if strings.Contains(tab.URL, "/sorry/") || strings.Contains(tab.URL, "about:blank") {
			continue
		}
		ctx, cancel := chromedp.NewContext(bridge.browserCtx)
		tCtx, tCancel := context.WithTimeout(ctx, 10*time.Second)
		if err := chromedp.Run(tCtx, chromedp.Navigate(tab.URL)); err != nil {
			tCancel()
			cancel()
			log.Printf("Failed to restore tab %s: %v", tab.URL, err)
			continue
		}
		tCancel()
		newID := string(chromedp.FromContext(ctx).Target.TargetID)
		bridge.tabs[newID] = &TabEntry{ctx: ctx, cancel: cancel}
		restored++
	}
	if restored > 0 {
		log.Printf("Restored %d tabs from previous session", restored)
	}
}
