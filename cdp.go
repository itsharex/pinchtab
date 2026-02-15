package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/chromedp/cdproto/target"
	"github.com/chromedp/chromedp"
)

func clickByNodeID(ctx context.Context, backendNodeID int64) error {
	return chromedp.Run(ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			p := map[string]any{"backendNodeId": backendNodeID}
			var result json.RawMessage
			if err := chromedp.FromContext(ctx).Target.Execute(ctx, "DOM.resolveNode", p, &result); err != nil {
				return fmt.Errorf("DOM.resolveNode: %w", err)
			}
			var resp struct {
				Object struct {
					ObjectID string `json:"objectId"`
				} `json:"object"`
			}
			if err := json.Unmarshal(result, &resp); err != nil {
				return fmt.Errorf("unmarshal resolveNode: %w", err)
			}
			if resp.Object.ObjectID == "" {
				return fmt.Errorf("no objectId for node %d", backendNodeID)
			}
			callP := map[string]any{
				"objectId":            resp.Object.ObjectID,
				"functionDeclaration": "function() { this.click(); }",
				"arguments":           []any{},
			}
			if err := chromedp.FromContext(ctx).Target.Execute(ctx, "Runtime.callFunctionOn", callP, nil); err != nil {
				return fmt.Errorf("click callFunctionOn: %w", err)
			}
			return nil
		}),
	)
}

func typeByNodeID(ctx context.Context, backendNodeID int64, text string) error {
	return chromedp.Run(ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			p := map[string]any{"backendNodeId": backendNodeID}
			if err := chromedp.FromContext(ctx).Target.Execute(ctx, "DOM.focus", p, nil); err != nil {
				return fmt.Errorf("DOM.focus: %w", err)
			}
			return nil
		}),
		chromedp.KeyEvent(text),
	)
}

func listTargets() ([]*target.Info, error) {
	var targets []*target.Info
	if err := chromedp.Run(bridge.browserCtx,
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
		if t.Type == "page" {
			pages = append(pages, t)
		}
	}
	return pages, nil
}
