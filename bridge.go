package main

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	cdp "github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/target"
	"github.com/chromedp/chromedp"
)

type TabEntry struct {
	ctx    context.Context
	cancel context.CancelFunc
}

type refCache struct {
	refs  map[string]int64
	nodes []A11yNode
}

type Bridge struct {
	allocCtx      context.Context
	browserCtx    context.Context
	tabs          map[string]*TabEntry
	snapshots     map[string]*refCache
	stealthScript string
	actions       map[string]ActionFunc
	locks         *lockManager
	mu            sync.RWMutex
}

func (b *Bridge) injectStealth(ctx context.Context) {
	if b.stealthScript == "" {
		return
	}
	if err := chromedp.Run(ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			_, err := page.AddScriptToEvaluateOnNewDocument(b.stealthScript).Do(ctx)
			return err
		}),
	); err != nil {
		slog.Warn("stealth injection failed", "err", err)
	}
}

func (b *Bridge) TabContext(tabID string) (context.Context, string, error) {
	if tabID == "" {
		targets, err := b.ListTargets()
		if err != nil {
			return nil, "", fmt.Errorf("list targets: %w", err)
		}
		if len(targets) == 0 {
			return nil, "", fmt.Errorf("no tabs open")
		}
		tabID = string(targets[0].TargetID)
	}

	b.mu.RLock()
	if entry, ok := b.tabs[tabID]; ok && entry.ctx != nil {
		b.mu.RUnlock()
		return entry.ctx, tabID, nil
	}
	b.mu.RUnlock()

	b.mu.Lock()
	defer b.mu.Unlock()

	if entry, ok := b.tabs[tabID]; ok && entry.ctx != nil {
		return entry.ctx, tabID, nil
	}

	if b.browserCtx == nil {
		return nil, "", fmt.Errorf("no browser connection")
	}

	ctx, cancel := chromedp.NewContext(b.browserCtx,
		chromedp.WithTargetID(target.ID(tabID)),
	)
	if err := chromedp.Run(ctx); err != nil {
		cancel()
		return nil, "", fmt.Errorf("tab %s not found: %w", tabID, err)
	}

	b.injectStealth(ctx)
	if cfg.NoAnimations {
		b.injectNoAnimations(ctx)
	}

	b.tabs[tabID] = &TabEntry{ctx: ctx, cancel: cancel}
	return ctx, tabID, nil
}

func (b *Bridge) CleanStaleTabs(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}

		targets, err := b.ListTargets()
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
				slog.Info("cleaned stale tab", "id", id)
			}
		}
		b.mu.Unlock()
	}
}

func (b *Bridge) GetRefCache(tabID string) *refCache {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.snapshots[tabID]
}

func (b *Bridge) SetRefCache(tabID string, cache *refCache) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.snapshots[tabID] = cache
}

func (b *Bridge) DeleteRefCache(tabID string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.snapshots, tabID)
}

func (b *Bridge) CreateTab(url string) (string, context.Context, context.CancelFunc, error) {
	if b.browserCtx == nil {
		return "", nil, nil, fmt.Errorf("no browser context available")
	}
	ctx, cancel := chromedp.NewContext(b.browserCtx)

	b.injectStealth(ctx)

	if cfg.NoAnimations {
		b.injectNoAnimations(ctx)
	}

	if cfg.BlockMedia {
		_ = setResourceBlocking(ctx, mediaBlockPatterns)
	} else if cfg.BlockImages {
		_ = setResourceBlocking(ctx, imageBlockPatterns)
	}

	navURL := "about:blank"
	if url != "" {
		navURL = url
	}
	if err := navigatePage(ctx, navURL); err != nil {
		cancel()
		return "", nil, nil, fmt.Errorf("new tab: %w", err)
	}

	newTargetID := string(chromedp.FromContext(ctx).Target.TargetID)
	b.mu.Lock()
	b.tabs[newTargetID] = &TabEntry{ctx: ctx, cancel: cancel}
	b.mu.Unlock()

	return newTargetID, ctx, cancel, nil
}

func (b *Bridge) CloseTab(tabID string) error {
	b.mu.Lock()
	entry, tracked := b.tabs[tabID]
	b.mu.Unlock()

	if tracked && entry.cancel != nil {
		entry.cancel()
	}

	closeCtx, closeCancel := context.WithTimeout(b.browserCtx, 5*time.Second)
	defer closeCancel()

	if err := target.CloseTarget(target.ID(tabID)).Do(cdp.WithExecutor(closeCtx, chromedp.FromContext(closeCtx).Browser)); err != nil {

		if !tracked {
			return nil
		}
		slog.Debug("close target CDP", "tabId", tabID, "err", err)
	}

	b.mu.Lock()
	delete(b.tabs, tabID)
	delete(b.snapshots, tabID)
	b.mu.Unlock()

	return nil
}

func (b *Bridge) ListTargets() ([]*target.Info, error) {
	if b.browserCtx == nil {
		return nil, fmt.Errorf("no browser connection")
	}
	var targets []*target.Info
	if err := chromedp.Run(b.browserCtx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			targets, err = target.GetTargets().Do(ctx)
			return err
		}),
	); err != nil {
		return nil, fmt.Errorf("get targets: %w", err)
	}

	pages := make([]*target.Info, 0)
	for _, t := range targets {
		if t.Type == targetTypePage {
			pages = append(pages, t)
		}
	}
	return pages, nil
}
