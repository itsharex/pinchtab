package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/target"
	"github.com/chromedp/chromedp"
)

// ── GET /health ────────────────────────────────────────────

func handleHealth(w http.ResponseWriter, r *http.Request) {
	targets, err := listTargets()
	if err != nil {
		jsonResp(w, 200, map[string]any{"status": "disconnected", "error": err.Error(), "cdp": cdpURL})
		return
	}
	jsonResp(w, 200, map[string]any{"status": "ok", "tabs": len(targets), "cdp": cdpURL})
}

// ── GET /tabs ──────────────────────────────────────────────

func handleTabs(w http.ResponseWriter, r *http.Request) {
	targets, err := listTargets()
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

// ── GET /snapshot ──────────────────────────────────────────

func handleSnapshot(w http.ResponseWriter, r *http.Request) {
	tabID := r.URL.Query().Get("tabId")
	filter := r.URL.Query().Get("filter")
	maxDepthStr := r.URL.Query().Get("depth")
	maxDepth := -1
	if maxDepthStr != "" {
		if d, err := strconv.Atoi(maxDepthStr); err == nil {
			maxDepth = d
		}
	}

	ctx, resolvedTabID, err := tabContext(tabID)
	if err != nil {
		jsonErr(w, 404, err)
		return
	}

	tCtx, tCancel := context.WithTimeout(ctx, actionTimeout)
	defer tCancel()
	go cancelOnClientDone(r.Context(), tCancel)

	var rawResult json.RawMessage
	if err := chromedp.Run(tCtx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			return chromedp.FromContext(ctx).Target.Execute(ctx,
				"Accessibility.getFullAXTree", nil, &rawResult)
		}),
	); err != nil {
		jsonErr(w, 500, fmt.Errorf("a11y tree: %w", err))
		return
	}

	var treeResp struct {
		Nodes []rawAXNode `json:"nodes"`
	}
	if err := json.Unmarshal(rawResult, &treeResp); err != nil {
		jsonErr(w, 500, fmt.Errorf("parse a11y tree: %w", err))
		return
	}

	// Build parent map for depth
	parentMap := make(map[string]string)
	for _, n := range treeResp.Nodes {
		for _, childID := range n.ChildIDs {
			parentMap[childID] = n.NodeID
		}
	}
	depthOf := func(nodeID string) int {
		d := 0
		cur := nodeID
		for {
			p, ok := parentMap[cur]
			if !ok {
				break
			}
			d++
			cur = p
		}
		return d
	}

	flat := make([]A11yNode, 0)
	refs := make(map[string]int64)
	refID := 0

	for _, n := range treeResp.Nodes {
		if n.Ignored {
			continue
		}

		role := n.Role.String()
		name := n.Name.String()

		if role == "none" || role == "generic" || role == "InlineTextBox" {
			continue
		}
		if name == "" && role == "StaticText" {
			continue
		}

		depth := depthOf(n.NodeID)
		if maxDepth >= 0 && depth > maxDepth {
			continue
		}
		if filter == "interactive" && !interactiveRoles[role] {
			continue
		}

		ref := fmt.Sprintf("e%d", refID)
		entry := A11yNode{
			Ref:   ref,
			Role:  role,
			Name:  name,
			Depth: depth,
		}

		if v := n.Value.String(); v != "" {
			entry.Value = v
		}
		if n.BackendDOMNodeID != 0 {
			entry.NodeID = n.BackendDOMNodeID
			refs[ref] = n.BackendDOMNodeID
		}

		for _, prop := range n.Properties {
			if prop.Name == "disabled" && prop.Value.String() == "true" {
				entry.Disabled = true
			}
			if prop.Name == "focused" && prop.Value.String() == "true" {
				entry.Focused = true
			}
		}

		flat = append(flat, entry)
		refID++
	}

	// Cache the ref→nodeID mapping for this tab
	bridge.mu.Lock()
	bridge.snapshots[resolvedTabID] = &refCache{refs: refs}
	bridge.mu.Unlock()

	var url, title string
	chromedp.Run(tCtx,
		chromedp.Location(&url),
		chromedp.Title(&title),
	)

	jsonResp(w, 200, map[string]any{
		"url":   url,
		"title": title,
		"nodes": flat,
		"count": len(flat),
	})
}

// ── GET /screenshot ────────────────────────────────────────

func handleScreenshot(w http.ResponseWriter, r *http.Request) {
	tabID := r.URL.Query().Get("tabId")

	ctx, _, err := tabContext(tabID)
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

	if r.URL.Query().Get("raw") == "true" {
		w.Header().Set("Content-Type", "image/jpeg")
		if _, err := w.Write(buf); err != nil {
			log.Printf("screenshot write: %v", err)
		}
		return
	}

	jsonResp(w, 200, map[string]any{
		"format": "jpeg",
		"base64": buf,
	})
}

// ── GET /text ──────────────────────────────────────────────

func handleText(w http.ResponseWriter, r *http.Request) {
	tabID := r.URL.Query().Get("tabId")

	ctx, _, err := tabContext(tabID)
	if err != nil {
		jsonErr(w, 404, err)
		return
	}

	tCtx, tCancel := context.WithTimeout(ctx, actionTimeout)
	defer tCancel()
	go cancelOnClientDone(r.Context(), tCancel)

	var text string
	if err := chromedp.Run(tCtx,
		chromedp.Evaluate(`document.body.innerText`, &text),
	); err != nil {
		jsonErr(w, 500, fmt.Errorf("text extract: %w", err))
		return
	}

	var url, title string
	chromedp.Run(tCtx,
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

func handleNavigate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TabID string `json:"tabId"`
		URL   string `json:"url"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodySize)).Decode(&req); err != nil {
		jsonErr(w, 400, fmt.Errorf("decode: %w", err))
		return
	}
	if req.URL == "" {
		jsonResp(w, 400, map[string]string{"error": "url required"})
		return
	}

	ctx, resolvedTabID, err := tabContext(req.TabID)
	if err != nil {
		jsonErr(w, 404, err)
		return
	}

	tCtx, tCancel := context.WithTimeout(ctx, 30*time.Second)
	defer tCancel()
	go cancelOnClientDone(r.Context(), tCancel)

	if err := chromedp.Run(tCtx, chromedp.Navigate(req.URL)); err != nil {
		jsonErr(w, 500, fmt.Errorf("navigate: %w", err))
		return
	}

	// Invalidate snapshot cache — page changed
	bridge.mu.Lock()
	delete(bridge.snapshots, resolvedTabID)
	bridge.mu.Unlock()

	var url, title string
	chromedp.Run(tCtx,
		chromedp.Location(&url),
		chromedp.Title(&title),
	)

	jsonResp(w, 200, map[string]any{"url": url, "title": title})
}

// ── POST /action ───────────────────────────────────────────

func handleAction(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TabID    string `json:"tabId"`
		Kind     string `json:"kind"`
		Ref      string `json:"ref"`
		Selector string `json:"selector"`
		Text     string `json:"text"`
		Key      string `json:"key"`
		NodeID   int64  `json:"nodeId"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodySize)).Decode(&req); err != nil {
		jsonErr(w, 400, fmt.Errorf("decode: %w", err))
		return
	}

	ctx, resolvedTabID, err := tabContext(req.TabID)
	if err != nil {
		jsonErr(w, 404, err)
		return
	}

	tCtx, tCancel := context.WithTimeout(ctx, actionTimeout)
	defer tCancel()
	go cancelOnClientDone(r.Context(), tCancel)

	// Resolve ref to backendNodeID from cached snapshot
	if req.Ref != "" && req.NodeID == 0 && req.Selector == "" {
		bridge.mu.RLock()
		cache := bridge.snapshots[resolvedTabID]
		bridge.mu.RUnlock()
		if cache != nil {
			if nid, ok := cache.refs[req.Ref]; ok {
				req.NodeID = nid
			}
		}
		if req.NodeID == 0 {
			jsonResp(w, 400, map[string]string{
				"error": fmt.Sprintf("ref %s not found — take a /snapshot first", req.Ref),
			})
			return
		}
	}

	var sel string
	if req.Selector != "" {
		sel = req.Selector
	}

	var result map[string]any

	switch req.Kind {
	case "click":
		if sel != "" {
			err = chromedp.Run(tCtx, chromedp.Click(sel, chromedp.ByQuery))
		} else if req.NodeID > 0 {
			err = clickByNodeID(tCtx, req.NodeID)
		} else {
			jsonResp(w, 400, map[string]string{"error": "need selector, ref, or nodeId"})
			return
		}
		result = map[string]any{"clicked": true}

	case "type":
		if req.Text == "" {
			jsonResp(w, 400, map[string]string{"error": "text required for type"})
			return
		}
		if sel != "" {
			err = chromedp.Run(tCtx,
				chromedp.Click(sel, chromedp.ByQuery),
				chromedp.SendKeys(sel, req.Text, chromedp.ByQuery),
			)
		} else if req.NodeID > 0 {
			err = typeByNodeID(tCtx, req.NodeID, req.Text)
		} else {
			jsonResp(w, 400, map[string]string{"error": "need selector or ref"})
			return
		}
		result = map[string]any{"typed": req.Text}

	case "fill":
		if sel != "" {
			err = chromedp.Run(tCtx,
				chromedp.SetValue(sel, req.Text, chromedp.ByQuery),
			)
		}
		result = map[string]any{"filled": req.Text}

	case "press":
		if req.Key == "" {
			jsonResp(w, 400, map[string]string{"error": "key required for press"})
			return
		}
		err = chromedp.Run(tCtx, chromedp.KeyEvent(req.Key))
		result = map[string]any{"pressed": req.Key}

	case "focus":
		if sel != "" {
			err = chromedp.Run(tCtx, chromedp.Focus(sel, chromedp.ByQuery))
		} else if req.NodeID > 0 {
			err = chromedp.Run(tCtx,
				chromedp.ActionFunc(func(ctx context.Context) error {
					p := map[string]any{"backendNodeId": req.NodeID}
					return chromedp.FromContext(ctx).Target.Execute(ctx, "DOM.focus", p, nil)
				}),
			)
		}
		result = map[string]any{"focused": true}

	default:
		jsonResp(w, 400, map[string]string{"error": fmt.Sprintf("unknown action: %s", req.Kind)})
		return
	}

	if err != nil {
		jsonErr(w, 500, fmt.Errorf("action %s: %w", req.Kind, err))
		return
	}

	jsonResp(w, 200, result)
}

// ── POST /evaluate ─────────────────────────────────────────

func handleEvaluate(w http.ResponseWriter, r *http.Request) {
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

	ctx, _, err := tabContext(req.TabID)
	if err != nil {
		jsonErr(w, 404, err)
		return
	}

	tCtx, tCancel := context.WithTimeout(ctx, actionTimeout)
	defer tCancel()
	go cancelOnClientDone(r.Context(), tCancel)

	var result any
	if err := chromedp.Run(tCtx,
		chromedp.Evaluate(req.Expression, &result),
	); err != nil {
		jsonErr(w, 500, fmt.Errorf("evaluate: %w", err))
		return
	}

	jsonResp(w, 200, map[string]any{"result": result})
}

// ── POST /tab ──────────────────────────────────────────────

func handleTab(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Action string `json:"action"` // "new" or "close"
		TabID  string `json:"tabId"`
		URL    string `json:"url"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodySize)).Decode(&req); err != nil {
		jsonErr(w, 400, fmt.Errorf("decode: %w", err))
		return
	}

	switch req.Action {
	case "new":
		ctx, cancel := chromedp.NewContext(bridge.browserCtx)

		url := "about:blank"
		if req.URL != "" {
			url = req.URL
		}
		if err := chromedp.Run(ctx, chromedp.Navigate(url)); err != nil {
			cancel()
			jsonErr(w, 500, fmt.Errorf("new tab: %w", err))
			return
		}

		newTargetID := string(chromedp.FromContext(ctx).Target.TargetID)
		bridge.mu.Lock()
		bridge.tabs[newTargetID] = &TabEntry{ctx: ctx, cancel: cancel}
		bridge.mu.Unlock()

		var curURL, title string
		chromedp.Run(ctx, chromedp.Location(&curURL), chromedp.Title(&title))
		jsonResp(w, 200, map[string]any{"tabId": newTargetID, "url": curURL, "title": title})

	case "close":
		if req.TabID == "" {
			jsonResp(w, 400, map[string]string{"error": "tabId required"})
			return
		}

		bridge.mu.Lock()
		if entry, ok := bridge.tabs[req.TabID]; ok {
			if entry.cancel != nil {
				entry.cancel()
			}
			delete(bridge.tabs, req.TabID)
			delete(bridge.snapshots, req.TabID)
		}
		bridge.mu.Unlock()

		ctx, cancel := chromedp.NewContext(bridge.browserCtx,
			chromedp.WithTargetID(target.ID(req.TabID)),
		)
		defer cancel()
		if err := chromedp.Run(ctx, page.Close()); err != nil {
			jsonErr(w, 500, fmt.Errorf("close tab: %w", err))
			return
		}
		jsonResp(w, 200, map[string]any{"closed": true})

	default:
		jsonResp(w, 400, map[string]string{"error": "action must be 'new' or 'close'"})
	}
}
