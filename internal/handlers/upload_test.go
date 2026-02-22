package handlers

import (
	"testing"
)

func TestDecodeFileData_DataURL(t *testing.T) {
	// 1x1 red PNG as data URL
	input := "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8/5+hHgAHggJ/PchI7wAAAABJRU5ErkJggg=="
	data, ext, err := decodeFileData(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ext != ".png" {
		t.Errorf("expected .png, got %s", ext)
	}
	if len(data) == 0 {
		t.Error("expected non-empty data")
	}
	// Check PNG magic bytes
	if data[0] != 0x89 || data[1] != 'P' {
		t.Error("expected PNG magic bytes")
	}
}

func TestDecodeFileData_RawBase64(t *testing.T) {
	// 1x1 red PNG as raw base64
	input := "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8/5+hHgAHggJ/PchI7wAAAABJRU5ErkJggg=="
	data, ext, err := decodeFileData(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ext != ".png" {
		t.Errorf("expected .png (sniffed), got %s", ext)
	}
	if len(data) == 0 {
		t.Error("expected non-empty data")
	}
}

func TestDecodeFileData_InvalidBase64(t *testing.T) {
	_, _, err := decodeFileData("not-valid-base64!!!")
	if err == nil {
		t.Error("expected error for invalid base64")
	}
}

func TestMimeToExt(t *testing.T) {
	tests := []struct {
		mime string
		ext  string
	}{
		{"image/png", ".png"},
		{"image/jpeg", ".jpg"},
		{"image/gif", ".gif"},
		{"image/webp", ".webp"},
		{"application/pdf", ".pdf"},
		{"text/plain", ".txt"},
		{"application/octet-stream", ".bin"},
	}
	for _, tt := range tests {
		if got := mimeToExt(tt.mime); got != tt.ext {
			t.Errorf("mimeToExt(%q) = %q, want %q", tt.mime, got, tt.ext)
		}
	}
}

func TestSniffExt(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		ext  string
	}{
		{"png", []byte{0x89, 'P', 'N', 'G'}, ".png"},
		{"jpg", []byte{0xFF, 0xD8, 0x00, 0x00}, ".jpg"},
		{"gif", []byte("GIF89a"), ".gif"},
		{"pdf", []byte("%PDF-1.4"), ".pdf"},
		{"unknown", []byte{0x00, 0x01, 0x02, 0x03}, ".bin"},
		{"short", []byte{0x00}, ".bin"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sniffExt(tt.data); got != tt.ext {
				t.Errorf("sniffExt() = %q, want %q", got, tt.ext)
			}
		})
	}
}
