package orchestrator

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pinchtab/pinchtab/internal/bridge"
)

func startLocalHTTPServer(t *testing.T, h http.Handler) (*httptest.Server, string) {
	t.Helper()
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	server := httptest.NewUnstartedServer(h)
	server.Listener = ln
	server.Start()
	return server, fmt.Sprintf("%d", ln.Addr().(*net.TCPAddr).Port)
}

func TestProxyTabRequest_FallsBackToOnlyRunningInstance(t *testing.T) {
	backend, port := startLocalHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"ok":true}`)
	}))
	defer backend.Close()

	o := NewOrchestrator(t.TempDir())
	o.client = backend.Client()
	o.instances["inst_1"] = &InstanceInternal{
		Instance: bridge.Instance{ID: "inst_1", Status: "running", Port: port},
		URL:      "http://localhost:" + port,
		cmd:      &mockCmd{pid: 1234, isAlive: true},
	}

	req := httptest.NewRequest(http.MethodGet, "/tabs/ABC123/snapshot", nil)
	req.SetPathValue("id", "ABC123")
	w := httptest.NewRecorder()

	orig := processAliveFunc
	processAliveFunc = func(pid int) bool { return true }
	defer func() { processAliveFunc = orig }()

	o.proxyTabRequest(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestSingleRunningInstance_MultipleInstancesReturnsNil(t *testing.T) {
	o := NewOrchestrator(t.TempDir())
	o.instances["inst_1"] = &InstanceInternal{Instance: bridge.Instance{ID: "inst_1", Status: "running"}, cmd: &mockCmd{pid: 1, isAlive: true}}
	o.instances["inst_2"] = &InstanceInternal{Instance: bridge.Instance{ID: "inst_2", Status: "running"}, cmd: &mockCmd{pid: 2, isAlive: true}}

	orig := processAliveFunc
	processAliveFunc = func(pid int) bool { return true }
	defer func() { processAliveFunc = orig }()

	if got := o.singleRunningInstance(); got != nil {
		t.Fatalf("expected nil, got %v", got.ID)
	}
}

func TestSingleRunningInstance_IgnoresStopped(t *testing.T) {
	o := NewOrchestrator(t.TempDir())
	o.instances["inst_1"] = &InstanceInternal{Instance: bridge.Instance{ID: "inst_1", Status: "running"}, cmd: &mockCmd{pid: 1, isAlive: true}}
	o.instances["inst_2"] = &InstanceInternal{Instance: bridge.Instance{ID: "inst_2", Status: "stopped"}}

	orig := processAliveFunc
	processAliveFunc = func(pid int) bool { return pid == 1 }
	defer func() { processAliveFunc = orig }()

	got := o.singleRunningInstance()
	if got == nil || got.ID != "inst_1" {
		t.Fatalf("got %#v, want inst_1", got)
	}
}
