package utils

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestTruncateStringToMaxUTF8Bytes(t *testing.T) {
	water := "水" // 3 UTF-8 bytes, 1 UTF-16 code unit
	tests := []struct {
		name     string
		s        string
		maxBytes int
		want     string
	}{
		{"max zero", "hello", 0, ""},
		{"empty", "", 10, ""},
		{"fits ASCII", "hello", 10, "hello"},
		{"truncate ASCII", strings.Repeat("a", 200), 5, "aaaaa"},
		{"CJK", strings.Repeat(water, 50), 100, strings.Repeat(water, 33)},
		{"emoji too big", "\U0001f600", 3, ""},
		{"emoji fits", "\U0001f600", 4, "\U0001f600"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TruncateStringToMaxUTF8Bytes(tt.s, tt.maxBytes)
			if got != tt.want {
				t.Errorf("TruncateStringToMaxUTF8Bytes(%q, %d) = %q, want %q", tt.s, tt.maxBytes, got, tt.want)
			}
			if len(got) > tt.maxBytes && tt.maxBytes > 0 {
				t.Errorf("len %d > maxBytes %d", len(got), tt.maxBytes)
			}
			if !utf8.ValidString(got) {
				t.Errorf("invalid UTF-8: %q", got)
			}
		})
	}
}
