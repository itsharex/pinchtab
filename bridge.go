package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/chromedp/cdproto/target"
	"github.com/chromedp/chromedp"
)

// TabEntry holds a chromedp context for an open tab.
type TabEntry struct {
	ctx    context.Context
	cancel context.CancelFunc
}

// refCache stores the ref→backendNodeID mapping from the last snapshot per tab.
// This avoids re-fetching the a11y tree on every action — refs stay stable
// until the next snapshot call.
type refCache struct {
	refs map[string]int64 // "e0" → backendNodeID
}

// Bridge is the central state holder for tab contexts and snapshot caches.
type Bridge struct {
	allocCtx   context.Context
	browserCtx context.Context // persistent browser context
	tabs       map[string]*TabEntry
	snapshots  map[string]*refCache // tabID → last snapshot's ref mapping
	mu         sync.RWMutex
}

var bridge Bridge

// tabContext returns the chromedp context for a tab and the resolved tabID.
// If tabID is empty, uses the first page target.
// Uses RLock for reads, upgrades to Lock only when creating a new entry.
func tabContext(tabID string) (context.Context, string, error) {
	if tabID == "" {
		targets, err := listTargets()
		if err != nil {
			return nil, "", fmt.Errorf("list targets: %w", err)
		}
		if len(targets) == 0 {
			return nil, "", fmt.Errorf("no tabs open")
		}
		tabID = string(targets[0].TargetID)
	}

	// Fast path: read lock
	bridge.mu.RLock()
	if entry, ok := bridge.tabs[tabID]; ok {
		bridge.mu.RUnlock()
		return entry.ctx, tabID, nil
	}
	bridge.mu.RUnlock()

	// Slow path: write lock, double-check
	bridge.mu.Lock()
	defer bridge.mu.Unlock()

	if entry, ok := bridge.tabs[tabID]; ok {
		return entry.ctx, tabID, nil
	}

	ctx, cancel := chromedp.NewContext(bridge.browserCtx,
		chromedp.WithTargetID(target.ID(tabID)),
	)
	if err := chromedp.Run(ctx); err != nil {
		cancel()
		return nil, "", fmt.Errorf("tab %s not found: %w", tabID, err)
	}

	bridge.tabs[tabID] = &TabEntry{ctx: ctx, cancel: cancel}
	return ctx, tabID, nil
}

// cleanStaleTabs periodically removes tab entries whose targets no longer exist.
// Exits when ctx is cancelled.
func (b *Bridge) cleanStaleTabs(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}

		targets, err := listTargets()
		if err != nil {
			continue
		}

		alive := make(map[string]bool, len(targets))
		for _, t := range targets {
			alive[string(t.TargetID)] = true
		}

		b.mu.Lock()
		for id, entry := range b.tabs {
			if !alive[id] {
				if entry.cancel != nil {
					entry.cancel()
				}
				delete(b.tabs, id)
				delete(b.snapshots, id)
				log.Printf("Cleaned stale tab: %s", id)
			}
		}
		b.mu.Unlock()
	}
}
