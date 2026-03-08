package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/pinchtab/pinchtab/internal/bridge"
	"github.com/pinchtab/pinchtab/internal/semantic"
	"github.com/pinchtab/pinchtab/internal/web"
)

type findRequest struct {
	Query           string  `json:"query"`
	TabID           string  `json:"tabId,omitempty"`
	Threshold       float64 `json:"threshold,omitempty"`
	TopK            int     `json:"topK,omitempty"`
	LexicalWeight   float64 `json:"lexicalWeight,omitempty"`
	EmbeddingWeight float64 `json:"embeddingWeight,omitempty"`
	Explain         bool    `json:"explain,omitempty"`
}

type findResponse struct {
	BestRef      string                  `json:"best_ref"`
	Confidence   string                  `json:"confidence"`
	Score        float64                 `json:"score"`
	Matches      []semantic.ElementMatch `json:"matches"`
	Strategy     string                  `json:"strategy"`
	Threshold    float64                 `json:"threshold"`
	LatencyMs    int64                   `json:"latency_ms"`
	ElementCount int                     `json:"element_count"`
}

// HandleFind performs semantic element matching against the accessibility
// snapshot for a tab. If no cached snapshot exists, it is fetched
// automatically via the existing snapshot infrastructure.
//
// @Endpoint POST /find
// @Description Find elements by natural language query
//
// @Param query string body Natural language description of the element (required)
// @Param tabId string body Tab ID (optional, defaults to active tab)
// @Param threshold float body Minimum similarity score (optional, default: 0.3)
// @Param topK int body Maximum results to return (optional, default: 3)
//
// @Response 200 application/json Returns matched elements with scores and metrics
// @Response 400 application/json Missing query
// @Response 404 application/json Tab not found
// @Response 500 application/json Snapshot or matching error
func (h *Handlers) HandleFind(w http.ResponseWriter, r *http.Request) {
	if err := h.ensureChrome(); err != nil {
		web.Error(w, 500, fmt.Errorf("chrome initialization: %w", err))
		return
	}

	var req findRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodySize)).Decode(&req); err != nil {
		web.Error(w, 400, fmt.Errorf("decode: %w", err))
		return
	}

	if pathID := r.PathValue("id"); pathID != "" {
		req.TabID = pathID
	}

	req.Query = strings.TrimSpace(req.Query)
	if req.Query == "" {
		web.Error(w, 400, fmt.Errorf("missing required field 'query'"))
		return
	}
	if req.Threshold <= 0 {
		req.Threshold = 0.3
	}
	if req.TopK <= 0 {
		req.TopK = 3
	}

	// Resolve tab context to get the resolved ID for cache lookup.
	// Keep ctxTab so we can reuse it for CDP operations (e.g. auto-refresh).
	ctxTab, resolvedTabID, err := h.Bridge.TabContext(req.TabID)
	if err != nil {
		web.Error(w, 404, err)
		return
	}

	// Try cached snapshot first; auto-fetch if not available.
	nodes := h.resolveSnapshotNodes(resolvedTabID)
	if len(nodes) == 0 {
		// Auto-refresh: take a fresh snapshot via CDP using the context
		// obtained from the initial TabContext call (resolvedTabID is the
		// raw CDPID and cannot be passed to TabContext again).
		h.refreshRefCache(ctxTab, resolvedTabID)
		nodes = h.resolveSnapshotNodes(resolvedTabID)
	}
	if len(nodes) == 0 {
		web.Error(w, 500, fmt.Errorf("no elements found in snapshot for tab %s — navigate first", resolvedTabID))
		return
	}

	// Build descriptors from A11yNodes.
	descs := make([]semantic.ElementDescriptor, len(nodes))
	for i, n := range nodes {
		descs[i] = semantic.ElementDescriptor{
			Ref:   n.Ref,
			Role:  n.Role,
			Name:  n.Name,
			Value: n.Value,
		}
	}

	start := time.Now()
	result, err := h.Matcher.Find(r.Context(), req.Query, descs, semantic.FindOptions{
		Threshold:       req.Threshold,
		TopK:            req.TopK,
		LexicalWeight:   req.LexicalWeight,
		EmbeddingWeight: req.EmbeddingWeight,
		Explain:         req.Explain,
	})
	if err != nil {
		web.Error(w, 500, fmt.Errorf("matcher error: %w", err))
		return
	}

	resp := findResponse{
		BestRef:      result.BestRef,
		Confidence:   result.ConfidenceLabel(),
		Score:        result.BestScore,
		Matches:      result.Matches,
		Strategy:     result.Strategy,
		Threshold:    req.Threshold,
		LatencyMs:    time.Since(start).Milliseconds(),
		ElementCount: result.ElementCount,
	}
	if resp.Matches == nil {
		resp.Matches = []semantic.ElementMatch{}
	}

	// Cache intent for recovery: store the query + best-match descriptor
	// so the recovery engine can reconstruct a search if the ref goes stale.
	if result.BestRef != "" && h.Recovery != nil {
		var bestDesc semantic.ElementDescriptor
		for _, d := range descs {
			if d.Ref == result.BestRef {
				bestDesc = d
				break
			}
		}
		h.Recovery.RecordIntent(resolvedTabID, result.BestRef, semantic.IntentEntry{
			Query:      req.Query,
			Descriptor: bestDesc,
			Score:      result.BestScore,
			Confidence: resp.Confidence,
			Strategy:   result.Strategy,
			CachedAt:   time.Now(),
		})
	}

	web.JSON(w, 200, resp)
}

// resolveSnapshotNodes returns cached A11yNodes for the tab, or an empty
// slice if no cache is available. The caller should use refreshRefCache
// to auto-fetch a fresh snapshot via CDP when this returns nil.
func (h *Handlers) resolveSnapshotNodes(tabID string) []bridge.A11yNode {
	cache := h.Bridge.GetRefCache(tabID)
	if cache != nil && len(cache.Nodes) > 0 {
		return cache.Nodes
	}
	return nil
}
