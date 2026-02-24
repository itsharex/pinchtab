//go:build integration

package integration

import (
	"testing"
)

// UP1-UP11: File upload tests (MANUAL TESTS — require file:// URL support)
//
// These tests are disabled in CI due to headless Chrome limitations:
// - file:// URL navigation not supported in headless mode
// - File picker security restrictions
// - No display server for UI interactions
//
// For manual testing, see tests/manual/file-upload.md
//
// Commented out original tests below for reference:
/*

func TestUpload_SingleFile(t *testing.T) { ... }
func TestUpload_MultipleFiles(t *testing.T) { ... }
func TestUpload_DefaultSelector(t *testing.T) { ... }
func TestUpload_InvalidSelector(t *testing.T) { ... }
func TestUpload_MissingFiles(t *testing.T) { ... }
func TestUpload_FileNotFound(t *testing.T) { ... }
func TestUpload_BadJSON(t *testing.T) { ... }

*/
