//go:build integration

package integration

import (
	"testing"
)

// SS1: Basic screenshot
func TestScreenshot_Basic(t *testing.T) {
	navigate(t, "https://example.com")
	code, body := httpGet(t, "/screenshot")
	if code != 200 {
		t.Fatalf("expected 200, got %d", code)
	}
	// Default returns JSON with base64
	if len(body) < 100 {
		t.Error("screenshot response too small")
	}
}

// SS2: Raw screenshot (MANUAL TEST — requires proper display/CDP support)
// This test is disabled in CI due to headless Chrome limitations with raw screenshot output.
// See tests/manual/screenshot-raw.md for manual testing steps.
// func TestScreenshot_Raw(t *testing.T) {
// 	navigate(t, "https://example.com")
// 	code, body := httpGet(t, "/screenshot?raw=true")
// 	if code != 200 {
// 		t.Fatalf("expected 200, got %d", code)
// 	}
// 	// JPEG starts with FF D8
// 	if len(body) < 2 || body[0] != 0xFF || body[1] != 0xD8 {
// 		t.Error("expected raw JPEG data (FF D8 header)")
// 	}
// }
