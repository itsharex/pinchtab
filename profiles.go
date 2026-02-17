package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"os/exec"
)

// ---------------------------------------------------------------------------
// Profile Manager — manages named Chrome profiles under ~/.pinchtab/profiles/
// ---------------------------------------------------------------------------

type ProfileManager struct {
	baseDir string // e.g. ~/.pinchtab/profiles
	tracker *ActionTracker
	mu      sync.RWMutex
}

type ProfileInfo struct {
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	CreatedAt time.Time `json:"createdAt"`
	SizeMB    float64   `json:"sizeMB"`
	Source    string    `json:"source,omitempty"` // "imported", "created"
}

func NewProfileManager(baseDir string) *ProfileManager {
	os.MkdirAll(baseDir, 0755)
	return &ProfileManager{
		baseDir: baseDir,
		tracker: NewActionTracker(),
	}
}

// List returns all managed profiles.
func (pm *ProfileManager) List() ([]ProfileInfo, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	entries, err := os.ReadDir(pm.baseDir)
	if err != nil {
		return nil, err
	}

	var profiles []ProfileInfo
	// Directories that aren't profiles
	skip := map[string]bool{"bin": true, "profiles": true}
	for _, e := range entries {
		if !e.IsDir() || skip[e.Name()] {
			continue
		}
		info, err := pm.profileInfo(e.Name())
		if err != nil {
			continue
		}
		// Must have a Default/ subdirectory to be a valid Chrome profile
		if _, err := os.Stat(filepath.Join(pm.baseDir, e.Name(), "Default")); err != nil {
			continue
		}
		profiles = append(profiles, info)
	}
	sort.Slice(profiles, func(i, j int) bool { return profiles[i].Name < profiles[j].Name })
	return profiles, nil
}

func (pm *ProfileManager) profileInfo(name string) (ProfileInfo, error) {
	dir := filepath.Join(pm.baseDir, name)
	fi, err := os.Stat(dir)
	if err != nil {
		return ProfileInfo{}, err
	}

	size := dirSizeMB(dir)
	source := "created"
	if _, err := os.Stat(filepath.Join(dir, ".pinchtab-imported")); err == nil {
		source = "imported"
	}

	return ProfileInfo{
		Name:      name,
		Path:      dir,
		CreatedAt: fi.ModTime(),
		SizeMB:    size,
		Source:    source,
	}, nil
}

// Import copies an existing Chrome user data directory as a named profile.
func (pm *ProfileManager) Import(name, sourcePath string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	dest := filepath.Join(pm.baseDir, name)
	if _, err := os.Stat(dest); err == nil {
		return fmt.Errorf("profile %q already exists", name)
	}

	// Validate source looks like a Chrome profile
	if _, err := os.Stat(filepath.Join(sourcePath, "Default")); err != nil {
		// Maybe it IS the Default folder
		if _, err2 := os.Stat(filepath.Join(sourcePath, "Preferences")); err2 != nil {
			return fmt.Errorf("source doesn't look like a Chrome user data dir (no Default/ or Preferences found)")
		}
	}

	slog.Info("importing profile", "name", name, "source", sourcePath)
	if err := exec.Command("cp", "-a", sourcePath, dest).Run(); err != nil {
		return fmt.Errorf("copy failed: %w", err)
	}

	// Mark as imported
	os.WriteFile(filepath.Join(dest, ".pinchtab-imported"), []byte(sourcePath), 0644)
	return nil
}

// Create makes a fresh empty profile directory.
func (pm *ProfileManager) Create(name string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	dest := filepath.Join(pm.baseDir, name)
	if _, err := os.Stat(dest); err == nil {
		return fmt.Errorf("profile %q already exists", name)
	}
	return os.MkdirAll(filepath.Join(dest, "Default"), 0755)
}

// Reset clears session data and caches but preserves bookmarks, extensions, passwords.
func (pm *ProfileManager) Reset(name string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	dir := filepath.Join(pm.baseDir, name)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("profile %q not found", name)
	}

	// Directories to nuke (session/cache data)
	nukeDirs := []string{
		"Default/Sessions",
		"Default/Session Storage",
		"Default/Cache",
		"Default/Code Cache",
		"Default/GPUCache",
		"Default/Service Worker",
		"Default/blob_storage",
		"ShaderCache",
		"GrShaderCache",
	}

	nukeFiles := []string{
		"Default/Cookies",
		"Default/Cookies-journal",
		"Default/History",
		"Default/History-journal",
		"Default/Visited Links",
	}

	for _, d := range nukeDirs {
		p := filepath.Join(dir, d)
		if err := os.RemoveAll(p); err != nil {
			slog.Warn("reset: failed to remove dir", "path", p, "err", err)
		}
	}
	for _, f := range nukeFiles {
		p := filepath.Join(dir, f)
		os.Remove(p)
	}

	slog.Info("profile reset", "name", name)
	return nil
}

// Delete removes a profile entirely.
func (pm *ProfileManager) Delete(name string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	dir := filepath.Join(pm.baseDir, name)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("profile %q not found", name)
	}
	return os.RemoveAll(dir)
}

// Logs returns recent Chrome/Pinchtab logs for a profile.
func (pm *ProfileManager) Logs(name string, limit int) []ActionRecord {
	return pm.tracker.GetLogs(name, limit)
}

// Analytics returns action patterns and optimization suggestions.
func (pm *ProfileManager) Analytics(name string) AnalyticsReport {
	return pm.tracker.Analyze(name)
}

// ---------------------------------------------------------------------------
// Action Tracker — records API calls per profile for pattern analysis
// ---------------------------------------------------------------------------

type ActionTracker struct {
	mu      sync.Mutex
	records map[string][]ActionRecord // profile -> records
}

type ActionRecord struct {
	Timestamp time.Time `json:"timestamp"`
	Method    string    `json:"method"`
	Endpoint  string    `json:"endpoint"`
	URL       string    `json:"url,omitempty"`    // target page URL if relevant
	TabID     string    `json:"tabId,omitempty"`
	DurationMs int64   `json:"durationMs"`
	Status    int       `json:"status"`
}

type AnalyticsReport struct {
	TotalActions    int                `json:"totalActions"`
	Since           time.Time          `json:"since"`
	TopEndpoints    []EndpointCount    `json:"topEndpoints"`
	RepeatPatterns  []RepeatPattern    `json:"repeatPatterns"`
	Suggestions     []string           `json:"suggestions"`
}

type EndpointCount struct {
	Endpoint string `json:"endpoint"`
	Count    int    `json:"count"`
	AvgMs    int64  `json:"avgMs"`
}

type RepeatPattern struct {
	Pattern   string `json:"pattern"`
	Count     int    `json:"count"`
	AvgGapSec float64 `json:"avgGapSec"`
}

func NewActionTracker() *ActionTracker {
	return &ActionTracker{
		records: make(map[string][]ActionRecord),
	}
}

func (at *ActionTracker) Record(profile string, rec ActionRecord) {
	at.mu.Lock()
	defer at.mu.Unlock()

	recs := at.records[profile]
	recs = append(recs, rec)
	// Keep last 10000 records per profile
	if len(recs) > 10000 {
		recs = recs[len(recs)-10000:]
	}
	at.records[profile] = recs
}

func (at *ActionTracker) GetLogs(profile string, limit int) []ActionRecord {
	at.mu.Lock()
	defer at.mu.Unlock()

	recs := at.records[profile]
	if limit <= 0 || limit > len(recs) {
		limit = len(recs)
	}
	// Return most recent
	start := len(recs) - limit
	result := make([]ActionRecord, limit)
	copy(result, recs[start:])
	return result
}

func (at *ActionTracker) Analyze(profile string) AnalyticsReport {
	at.mu.Lock()
	defer at.mu.Unlock()

	recs := at.records[profile]
	if len(recs) == 0 {
		return AnalyticsReport{Suggestions: []string{"No actions recorded yet."}}
	}

	report := AnalyticsReport{
		TotalActions: len(recs),
		Since:        recs[0].Timestamp,
	}

	// Top endpoints
	epCounts := map[string]struct{ count int; totalMs int64 }{}
	for _, r := range recs {
		v := epCounts[r.Endpoint]
		v.count++
		v.totalMs += r.DurationMs
		epCounts[r.Endpoint] = v
	}
	for ep, v := range epCounts {
		report.TopEndpoints = append(report.TopEndpoints, EndpointCount{
			Endpoint: ep,
			Count:    v.count,
			AvgMs:    v.totalMs / int64(v.count),
		})
	}
	sort.Slice(report.TopEndpoints, func(i, j int) bool {
		return report.TopEndpoints[i].Count > report.TopEndpoints[j].Count
	})
	if len(report.TopEndpoints) > 10 {
		report.TopEndpoints = report.TopEndpoints[:10]
	}

	// Detect repeated snapshot patterns (same URL polled frequently)
	urlSnaps := map[string][]time.Time{}
	for _, r := range recs {
		if r.Endpoint == "/snapshot" && r.URL != "" {
			urlSnaps[r.URL] = append(urlSnaps[r.URL], r.Timestamp)
		}
	}
	for url, times := range urlSnaps {
		if len(times) < 3 {
			continue
		}
		var totalGap float64
		for i := 1; i < len(times); i++ {
			totalGap += times[i].Sub(times[i-1]).Seconds()
		}
		avgGap := totalGap / float64(len(times)-1)
		report.RepeatPatterns = append(report.RepeatPatterns, RepeatPattern{
			Pattern:   fmt.Sprintf("snapshot %s", truncURL(url)),
			Count:     len(times),
			AvgGapSec: avgGap,
		})
	}

	// Detect repeated navigate→snapshot sequences
	navSnap := map[string]int{}
	for i := 1; i < len(recs); i++ {
		if recs[i-1].Endpoint == "/navigate" && recs[i].Endpoint == "/snapshot" {
			key := truncURL(recs[i-1].URL)
			navSnap[key]++
		}
	}
	for url, count := range navSnap {
		if count >= 3 {
			report.RepeatPatterns = append(report.RepeatPatterns, RepeatPattern{
				Pattern: fmt.Sprintf("navigate→snapshot %s", url),
				Count:   count,
			})
		}
	}

	// Generate suggestions
	for _, rp := range report.RepeatPatterns {
		if strings.HasPrefix(rp.Pattern, "snapshot") && rp.AvgGapSec > 0 && rp.AvgGapSec < 10 {
			report.Suggestions = append(report.Suggestions,
				fmt.Sprintf("High-frequency polling detected: %s every %.0fs — consider increasing interval or using smart diff", rp.Pattern, rp.AvgGapSec))
		}
		if strings.HasPrefix(rp.Pattern, "navigate→snapshot") && rp.Count > 5 {
			report.Suggestions = append(report.Suggestions,
				fmt.Sprintf("Repeated navigate→snapshot for same URL %s (%dx) — consider caching or using /text for lighter reads", rp.Pattern, rp.Count))
		}
	}

	// Check for heavy snapshot usage without selector/maxTokens
	snapCount := 0
	for _, r := range recs {
		if r.Endpoint == "/snapshot" {
			snapCount++
		}
	}
	if snapCount > 20 {
		report.Suggestions = append(report.Suggestions,
			fmt.Sprintf("Heavy snapshot usage (%d calls) — use ?selector= or ?maxTokens= to reduce token cost", snapCount))
	}

	if len(report.Suggestions) == 0 {
		report.Suggestions = []string{"No optimization suggestions — usage looks efficient."}
	}

	return report
}

// ---------------------------------------------------------------------------
// HTTP Handlers
// ---------------------------------------------------------------------------

func (pm *ProfileManager) RegisterHandlers(mux *http.ServeMux) {
	mux.HandleFunc("GET /profiles", pm.handleList)
	mux.HandleFunc("POST /profiles/import", pm.handleImport)
	mux.HandleFunc("POST /profiles/create", pm.handleCreate)
	mux.HandleFunc("POST /profiles/{name}/reset", pm.handleReset)
	mux.HandleFunc("DELETE /profiles/{name}", pm.handleDelete)
	mux.HandleFunc("GET /profiles/{name}/logs", pm.handleLogs)
	mux.HandleFunc("GET /profiles/{name}/analytics", pm.handleAnalytics)
}

func (pm *ProfileManager) handleList(w http.ResponseWriter, r *http.Request) {
	profiles, err := pm.List()
	if err != nil {
		jsonErr(w, http.StatusInternalServerError, err)
		return
	}
	jsonResp(w, http.StatusOK, profiles)
}

func (pm *ProfileManager) handleImport(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name   string `json:"name"`
		Source string `json:"source"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonErr(w, http.StatusBadRequest, fmt.Errorf("invalid JSON"))
		return
	}
	if req.Name == "" || req.Source == "" {
		jsonErr(w, http.StatusBadRequest, fmt.Errorf("name and source required"))
		return
	}
	if err := pm.Import(req.Name, req.Source); err != nil {
		jsonErr(w, http.StatusConflict, err)
		return
	}
	jsonResp(w, http.StatusCreated, map[string]string{"status": "imported", "name": req.Name})
}

func (pm *ProfileManager) handleCreate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonErr(w, http.StatusBadRequest, fmt.Errorf("invalid JSON"))
		return
	}
	if req.Name == "" {
		jsonErr(w, http.StatusBadRequest, fmt.Errorf("name required"))
		return
	}
	if err := pm.Create(req.Name); err != nil {
		jsonErr(w, http.StatusConflict, err)
		return
	}
	jsonResp(w, http.StatusCreated, map[string]string{"status": "created", "name": req.Name})
}

func (pm *ProfileManager) handleReset(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if err := pm.Reset(name); err != nil {
		jsonErr(w, http.StatusNotFound, err)
		return
	}
	jsonResp(w, http.StatusOK, map[string]string{"status": "reset", "name": name})
}

func (pm *ProfileManager) handleDelete(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if err := pm.Delete(name); err != nil {
		jsonErr(w, http.StatusNotFound, err)
		return
	}
	jsonResp(w, http.StatusOK, map[string]string{"status": "deleted", "name": name})
}

func (pm *ProfileManager) handleLogs(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	limit := profileQueryInt(r, "limit", 100)
	logs := pm.Logs(name, limit)
	jsonResp(w, http.StatusOK, logs)
}

func (pm *ProfileManager) handleAnalytics(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	report := pm.Analytics(name)
	jsonResp(w, http.StatusOK, report)
}

// TrackingMiddleware returns middleware that records actions per profile.
func (pm *ProfileManager) TrackingMiddleware(profileName string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, code: 200}
		next.ServeHTTP(sw, r)

		// Record the action
		rec := ActionRecord{
			Timestamp:  start,
			Method:     r.Method,
			Endpoint:   r.URL.Path,
			URL:        r.URL.Query().Get("url"),
			TabID:      r.URL.Query().Get("tabId"),
			DurationMs: time.Since(start).Milliseconds(),
			Status:     sw.code,
		}
		pm.tracker.Record(profileName, rec)
	})
}

// uses statusWriter from middleware.go
// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func profileQueryInt(r *http.Request, key string, def int) int {
	s := r.URL.Query().Get(key)
	if s == "" {
		return def
	}
	var n int
	fmt.Sscanf(s, "%d", &n)
	if n <= 0 {
		return def
	}
	return n
}

func dirSizeMB(path string) float64 {
	var total int64
	filepath.WalkDir(path, func(_ string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err == nil {
			total += info.Size()
		}
		return nil
	})
	return float64(total) / (1024 * 1024)
}

func truncURL(u string) string {
	if len(u) > 60 {
		return u[:57] + "..."
	}
	return u
}
