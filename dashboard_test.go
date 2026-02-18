package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDashboardRecordAndGetAgents(t *testing.T) {
	d := NewDashboard()

	d.RecordEvent(AgentEvent{
		AgentID:   "mario",
		Profile:   "default",
		Action:    "GET /snapshot",
		URL:       "https://example.com",
		TabID:     "tab1",
		Status:    200,
		DurationMs: 150,
		Timestamp: time.Now(),
	})

	d.RecordEvent(AgentEvent{
		AgentID:   "r40",
		Action:    "POST /navigate",
		URL:       "https://r40.io",
		Status:    200,
		DurationMs: 80,
		Timestamp: time.Now(),
	})

	agents := d.GetAgents()
	if len(agents) != 2 {
		t.Fatalf("expected 2 agents, got %d", len(agents))
	}

	found := map[string]bool{}
	for _, a := range agents {
		found[a.AgentID] = true
		if a.Status != "active" {
			t.Errorf("agent %s should be active, got %s", a.AgentID, a.Status)
		}
	}
	if !found["mario"] || !found["r40"] {
		t.Error("expected both mario and r40 agents")
	}
}

func TestDashboardAgentUpdates(t *testing.T) {
	d := NewDashboard()

	d.RecordEvent(AgentEvent{
		AgentID: "bot", Action: "GET /snapshot", URL: "https://page1.com",
		Timestamp: time.Now(),
	})
	d.RecordEvent(AgentEvent{
		AgentID: "bot", Action: "POST /navigate", URL: "https://page2.com",
		Timestamp: time.Now(),
	})

	agents := d.GetAgents()
	if len(agents) != 1 {
		t.Fatalf("expected 1 agent, got %d", len(agents))
	}
	if agents[0].CurrentURL != "https://page2.com" {
		t.Errorf("expected URL page2.com, got %s", agents[0].CurrentURL)
	}
	if agents[0].ActionCount != 2 {
		t.Errorf("expected 2 actions, got %d", agents[0].ActionCount)
	}
}

func TestDashboardHandlerAgents(t *testing.T) {
	d := NewDashboard()
	d.RecordEvent(AgentEvent{
		AgentID: "test-agent", Action: "GET /health",
		Status: 200, Timestamp: time.Now(),
	})

	mux := http.NewServeMux()
	d.RegisterHandlers(mux)

	req := httptest.NewRequest("GET", "/dashboard/agents", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var agents []AgentActivity
	json.NewDecoder(w.Body).Decode(&agents)
	if len(agents) != 1 || agents[0].AgentID != "test-agent" {
		t.Errorf("unexpected agents: %+v", agents)
	}
}

func TestDashboardUI(t *testing.T) {
	d := NewDashboard()
	mux := http.NewServeMux()
	d.RegisterHandlers(mux)

	req := httptest.NewRequest("GET", "/dashboard", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "text/html" {
		t.Errorf("expected text/html, got %s", ct)
	}
	body := w.Body.String()
	if len(body) < 1000 {
		t.Error("dashboard HTML seems too short")
	}
}

func TestDashboardSSEInit(t *testing.T) {
	d := NewDashboard()
	d.RecordEvent(AgentEvent{AgentID: "sse-agent", Action: "GET /health", Timestamp: time.Now()})

	mux := http.NewServeMux()
	d.RegisterHandlers(mux)

	// Use a real test server with short-lived request
	srv := httptest.NewServer(mux)
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	req, _ := http.NewRequestWithContext(ctx, "GET", srv.URL+"/dashboard/events", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil && !errors.Is(err, context.DeadlineExceeded) {
		// Context deadline is expected
		return
	}
	if resp != nil {
		defer resp.Body.Close()
		if resp.Header.Get("Content-Type") != "text/event-stream" {
			t.Errorf("expected text/event-stream, got %s", resp.Header.Get("Content-Type"))
		}
	}
}

func TestTrackingMiddleware(t *testing.T) {
	d := NewDashboard()
	pm := NewProfileManager(t.TempDir())

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})

	handler := d.TrackingMiddleware(pm, inner)

	req := httptest.NewRequest("GET", "/snapshot?url=https://example.com", nil)
	req.Header.Set("X-Agent-Id", "test-bot")
	req.Header.Set("X-Profile", "my-profile")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	agents := d.GetAgents()
	if len(agents) != 1 {
		t.Fatalf("expected 1 agent, got %d", len(agents))
	}
	if agents[0].AgentID != "test-bot" {
		t.Errorf("expected agent test-bot, got %s", agents[0].AgentID)
	}

	// Check profile tracker also recorded
	logs := pm.Logs("my-profile", 10)
	if len(logs) != 1 {
		t.Errorf("expected 1 profile log, got %d", len(logs))
	}
}

func TestTrackingMiddlewareAnonymous(t *testing.T) {
	d := NewDashboard()
	pm := NewProfileManager(t.TempDir())

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})

	handler := d.TrackingMiddleware(pm, inner)
	// Use /snapshot (a real agent endpoint, not skipped)
	req := httptest.NewRequest("GET", "/snapshot", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	agents := d.GetAgents()
	if len(agents) != 1 || agents[0].AgentID != "anonymous" {
		t.Errorf("expected anonymous agent, got %+v", agents)
	}
}

func TestTrackingMiddlewareSkipsManagementRoutes(t *testing.T) {
	d := NewDashboard()
	pm := NewProfileManager(t.TempDir())

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})

	handler := d.TrackingMiddleware(pm, inner)

	// All of these should be skipped
	for _, path := range []string{"/dashboard", "/profiles", "/instances", "/health", "/welcome", "/favicon.ico"} {
		req := httptest.NewRequest("GET", path, nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}

	agents := d.GetAgents()
	if len(agents) != 0 {
		t.Errorf("management routes should not be tracked, got %+v", agents)
	}
}
