package bridge

import (
	"context"
	"time"

	"github.com/chromedp/cdproto/performance"
	"github.com/chromedp/cdproto/target"
	"github.com/chromedp/chromedp"
)

// MemoryMetrics holds Chrome memory statistics
type MemoryMetrics struct {
	JSHeapUsedMB  float64 `json:"jsHeapUsedMB"`
	JSHeapTotalMB float64 `json:"jsHeapTotalMB"`
	Documents     int64   `json:"documents"`
	Frames        int64   `json:"frames"`
	Nodes         int64   `json:"nodes"`
	Listeners     int64   `json:"listeners"`
}

// GetMemoryMetrics retrieves memory metrics for a specific tab
func (b *Bridge) GetMemoryMetrics(tabID string) (*MemoryMetrics, error) {
	ctx, _, err := b.TabContext(tabID)
	if err != nil {
		return nil, err
	}

	return getMetricsFromContext(ctx)
}

// GetBrowserMemoryMetrics retrieves memory metrics for the entire browser
func (b *Bridge) GetBrowserMemoryMetrics() (*MemoryMetrics, error) {
	if b.BrowserCtx == nil {
		return nil, nil
	}
	return getMetricsFromContext(b.BrowserCtx)
}

// GetAggregatedMemoryMetrics returns summed memory metrics across all open tabs
func (b *Bridge) GetAggregatedMemoryMetrics() (*MemoryMetrics, error) {
	targets, err := b.ListTargets()
	if err != nil {
		return nil, err
	}

	total := &MemoryMetrics{}
	for _, t := range targets {
		if t.Type != "page" {
			continue
		}
		// Try to get metrics - skip on any error
		mem := b.safeGetMetricsForTarget(string(t.TargetID))
		if mem == nil {
			continue
		}
		total.JSHeapUsedMB += mem.JSHeapUsedMB
		total.JSHeapTotalMB += mem.JSHeapTotalMB
		total.Documents += mem.Documents
		total.Frames += mem.Frames
		total.Nodes += mem.Nodes
		total.Listeners += mem.Listeners
	}
	return total, nil
}

// safeGetMetricsForTarget safely gets metrics, returning nil on any error
func (b *Bridge) safeGetMetricsForTarget(targetID string) *MemoryMetrics {
	mem, err := b.getMetricsForTarget(targetID)
	if err != nil {
		return nil
	}
	return mem
}

// getMetricsForTarget gets metrics for a raw CDP target ID
func (b *Bridge) getMetricsForTarget(targetID string) (result *MemoryMetrics, err error) {
	// Recover from panics - some targets may not support metrics
	defer func() {
		if r := recover(); r != nil {
			result = nil
			err = nil // swallow error, just skip this target
		}
	}()

	// Create context with timeout to avoid hanging
	ctx, cancel := chromedp.NewContext(b.BrowserCtx, chromedp.WithTargetID(target.ID(targetID)))
	defer cancel()

	timeoutCtx, timeoutCancel := context.WithTimeout(ctx, 2*time.Second)
	defer timeoutCancel()

	return getMetricsFromContext(timeoutCtx)
}

func getMetricsFromContext(ctx context.Context) (*MemoryMetrics, error) {
	// Enable performance metrics collection
	if err := chromedp.Run(ctx, performance.Enable()); err != nil {
		return nil, err
	}

	var metrics []*performance.Metric
	if err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		var err error
		metrics, err = performance.GetMetrics().Do(ctx)
		return err
	})); err != nil {
		return nil, err
	}

	result := &MemoryMetrics{}
	for _, m := range metrics {
		switch m.Name {
		case "JSHeapUsedSize":
			result.JSHeapUsedMB = m.Value / (1024 * 1024)
		case "JSHeapTotalSize":
			result.JSHeapTotalMB = m.Value / (1024 * 1024)
		case "Documents":
			result.Documents = int64(m.Value)
		case "Frames":
			result.Frames = int64(m.Value)
		case "Nodes":
			result.Nodes = int64(m.Value)
		case "JSEventListeners":
			result.Listeners = int64(m.Value)
		}
	}

	return result, nil
}
