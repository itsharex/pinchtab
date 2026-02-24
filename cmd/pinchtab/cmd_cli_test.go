package main

import "testing"

func TestIsCLICommand(t *testing.T) {
	valid := []string{"nav", "navigate", "snap", "snapshot", "click", "type",
		"press", "fill", "hover", "scroll", "select", "focus",
		"text", "tabs", "tab", "screenshot", "ss", "eval", "evaluate",
		"pdf", "health"}

	for _, cmd := range valid {
		if !isCLICommand(cmd) {
			t.Errorf("expected %q to be a CLI command", cmd)
		}
	}

	invalid := []string{"dashboard", "connect", "config", "server", "run", ""}
	for _, cmd := range invalid {
		if isCLICommand(cmd) {
			t.Errorf("expected %q to NOT be a CLI command", cmd)
		}
	}
}

func TestPrintHelp(t *testing.T) {
	// Just verify it doesn't panic
	printHelp()
}
