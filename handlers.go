package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

// ── GET /health ────────────────────────────────────────────

func (b *Bridge) handleHealth(w http.ResponseWriter, r *http.Request) {
	targets, err := b.ListTargets()
	if err != nil {
		jsonResp(w, 200, map[string]any{"status": "disconnected", "error": err.Error(), "cdp": cdpURL})
		return
	}
	jsonResp(w, 200, map[string]any{"status": "ok", "tabs": len(targets), "cdp": cdpURL})
}

// ── GET /tabs ──────────────────────────────────────────────

func (b *Bridge) handleTabs(w http.ResponseWriter, r *http.Request) {
	targets, err := b.ListTargets()
	if err != nil {
		jsonErr(w, 500, err)
		return
	}

	tabs := make([]map[string]any, 0, len(targets))
	for _, t := range targets {
		tabs = append(tabs, map[string]any{
			"id":    string(t.TargetID),
			"url":   t.URL,
			"title": t.Title,
			"type":  t.Type,
		})
	}
	jsonResp(w, 200, map[string]any{"tabs": tabs})
}

// ── GET /screenshot ────────────────────────────────────────

func (b *Bridge) handleScreenshot(w http.ResponseWriter, r *http.Request) {
	tabID := r.URL.Query().Get("tabId")
	output := r.URL.Query().Get("output")

	ctx, _, err := b.TabContext(tabID)
	if err != nil {
		jsonErr(w, 404, err)
		return
	}

	tCtx, tCancel := context.WithTimeout(ctx, actionTimeout)
	defer tCancel()
	go cancelOnClientDone(r.Context(), tCancel)

	var buf []byte
	quality := 80
	if q := r.URL.Query().Get("quality"); q != "" {
		if qn, err := strconv.Atoi(q); err == nil {
			quality = qn
		}
	}

	if err := chromedp.Run(tCtx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			buf, err = page.CaptureScreenshot().
				WithFormat(page.CaptureScreenshotFormatJpeg).
				WithQuality(int64(quality)).
				Do(ctx)
			return err
		}),
	); err != nil {
		jsonErr(w, 500, fmt.Errorf("screenshot: %w", err))
		return
	}

	// Handle file output
	if output == "file" {
		screenshotDir := filepath.Join(stateDir, "screenshots")
		if err := os.MkdirAll(screenshotDir, 0755); err != nil {
			jsonErr(w, 500, fmt.Errorf("create screenshot dir: %w", err))
			return
		}

		timestamp := time.Now().Format("20060102-150405")
		filename := fmt.Sprintf("screenshot-%s.jpg", timestamp)
		filePath := filepath.Join(screenshotDir, filename)

		if err := os.WriteFile(filePath, buf, 0644); err != nil {
			jsonErr(w, 500, fmt.Errorf("write screenshot: %w", err))
			return
		}

		jsonResp(w, 200, map[string]any{
			"path":      filePath,
			"size":      len(buf),
			"format":    "jpeg",
			"timestamp": timestamp,
		})
		return
	}

	if r.URL.Query().Get("raw") == "true" {
		w.Header().Set("Content-Type", "image/jpeg")
		if _, err := w.Write(buf); err != nil {
			slog.Error("screenshot write", "err", err)
		}
		return
	}

	jsonResp(w, 200, map[string]any{
		"format": "jpeg",
		"base64": base64.StdEncoding.EncodeToString(buf),
	})
}

// ── GET /text ──────────────────────────────────────────────

func (b *Bridge) handleText(w http.ResponseWriter, r *http.Request) {
	tabID := r.URL.Query().Get("tabId")
	mode := r.URL.Query().Get("mode")

	ctx, _, err := b.TabContext(tabID)
	if err != nil {
		jsonErr(w, 404, err)
		return
	}

	tCtx, tCancel := context.WithTimeout(ctx, actionTimeout)
	defer tCancel()
	go cancelOnClientDone(r.Context(), tCancel)

	var text string
	if mode == "raw" {
		if err := chromedp.Run(tCtx,
			chromedp.Evaluate(`document.body.innerText`, &text),
		); err != nil {
			jsonErr(w, 500, fmt.Errorf("text extract: %w", err))
			return
		}
	} else {
		if err := chromedp.Run(tCtx,
			chromedp.Evaluate(readabilityJS, &text),
		); err != nil {
			jsonErr(w, 500, fmt.Errorf("text extract: %w", err))
			return
		}
	}

	var url, title string
	_ = chromedp.Run(tCtx,
		chromedp.Location(&url),
		chromedp.Title(&title),
	)

	jsonResp(w, 200, map[string]any{
		"url":   url,
		"title": title,
		"text":  text,
	})
}

// ── POST /navigate ─────────────────────────────────────────

func (b *Bridge) handleNavigate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TabID  string `json:"tabId"`
		URL    string `json:"url"`
		NewTab bool   `json:"newTab"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodySize)).Decode(&req); err != nil {
		jsonErr(w, 400, fmt.Errorf("decode: %w", err))
		return
	}
	if req.URL == "" {
		jsonResp(w, 400, map[string]string{"error": "url required"})
		return
	}

	if req.NewTab {
		newTargetID, newCtx, _, err := b.CreateTab(req.URL)
		if err != nil {
			jsonErr(w, 500, fmt.Errorf("new tab: %w", err))
			return
		}

		tCtx, tCancel := context.WithTimeout(newCtx, navigateTimeout)
		defer tCancel()
		go cancelOnClientDone(r.Context(), tCancel)

		var url, title string
		_ = chromedp.Run(tCtx, chromedp.Location(&url))
		title = waitForTitle(tCtx)

		jsonResp(w, 200, map[string]any{"tabId": newTargetID, "url": url, "title": title})
		return
	}

	ctx, resolvedTabID, err := b.TabContext(req.TabID)
	if err != nil {
		jsonErr(w, 404, err)
		return
	}

	tCtx, tCancel := context.WithTimeout(ctx, navigateTimeout)
	defer tCancel()
	go cancelOnClientDone(r.Context(), tCancel)

	if err := navigatePage(tCtx, req.URL); err != nil {
		jsonErr(w, 500, fmt.Errorf("navigate: %w", err))
		return
	}

	b.DeleteRefCache(resolvedTabID)

	var url string
	_ = chromedp.Run(tCtx, chromedp.Location(&url))
	title := waitForTitle(tCtx)

	jsonResp(w, 200, map[string]any{"url": url, "title": title})
}

// ── POST /evaluate ─────────────────────────────────────────

func (b *Bridge) handleEvaluate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TabID      string `json:"tabId"`
		Expression string `json:"expression"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodySize)).Decode(&req); err != nil {
		jsonErr(w, 400, fmt.Errorf("decode: %w", err))
		return
	}
	if req.Expression == "" {
		jsonResp(w, 400, map[string]string{"error": "expression required"})
		return
	}

	ctx, _, err := b.TabContext(req.TabID)
	if err != nil {
		jsonErr(w, 404, err)
		return
	}

	tCtx, tCancel := context.WithTimeout(ctx, actionTimeout)
	defer tCancel()
	go cancelOnClientDone(r.Context(), tCancel)

	var result any
	if err := chromedp.Run(tCtx, chromedp.Evaluate(req.Expression, &result)); err != nil {
		jsonErr(w, 500, fmt.Errorf("evaluate: %w", err))
		return
	}

	jsonResp(w, 200, map[string]any{"result": result})
}

// ── POST /tab ──────────────────────────────────────────────

func (b *Bridge) handleTab(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Action string `json:"action"`
		TabID  string `json:"tabId"`
		URL    string `json:"url"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodySize)).Decode(&req); err != nil {
		jsonErr(w, 400, fmt.Errorf("decode: %w", err))
		return
	}

	switch req.Action {
	case tabActionNew:
		newTargetID, ctx, _, err := b.CreateTab(req.URL)
		if err != nil {
			jsonErr(w, 500, err)
			return
		}

		var curURL, title string
		_ = chromedp.Run(ctx, chromedp.Location(&curURL), chromedp.Title(&title))
		jsonResp(w, 200, map[string]any{"tabId": newTargetID, "url": curURL, "title": title})

	case tabActionClose:
		if req.TabID == "" {
			jsonResp(w, 400, map[string]string{"error": "tabId required"})
			return
		}

		if err := b.CloseTab(req.TabID); err != nil {
			jsonErr(w, 500, err)
			return
		}
		jsonResp(w, 200, map[string]any{"closed": true})

	default:
		jsonResp(w, 400, map[string]string{"error": "action must be 'new' or 'close'"})
	}
}

// Shared helpers (jsonResp, jsonErr, cancelOnClientDone, waitForTitle)
// are defined in middleware.go and cdp.go respectively.
