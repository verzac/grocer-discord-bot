package utils

import (
	"strings"
	"testing"
)

func TestUTF16UnitsForRune(t *testing.T) {
	tests := []struct {
		name string
		r    rune
		want int
	}{
		{"ASCII", 'a', 1},
		{"BMP max", '\uffff', 1},
		{"supplementary plane (emoji)", '\U0001f600', 2},
		{"first supplementary", '\U00010000', 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := UTF16UnitsForRune(tt.r); got != tt.want {
				t.Errorf("UTF16UnitsForRune(%q) = %d, want %d", tt.r, got, tt.want)
			}
		})
	}
}

func TestUTF16Len(t *testing.T) {
	emoji := "\U0001f600" // 😀 — 2 UTF-16 code units in JS/Discord
	tests := []struct {
		name string
		s    string
		want int
	}{
		{"empty", "", 0},
		{"ASCII", "hello", 5},
		{"one emoji", emoji, 2},
		{"mixed", "a" + emoji + "b", 1 + 2 + 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := UTF16Len(tt.s); got != tt.want {
				t.Errorf("UTF16Len(%q) = %d, want %d", tt.s, got, tt.want)
			}
		})
	}
}

func TestTruncateStringForDiscord(t *testing.T) {
	emoji := "\U0001f600"
	longASCII := strings.Repeat("a", 200)

	tests := []struct {
		name     string
		s        string
		maxUTF16 int
		want     string
	}{
		{"max zero", "hello", 0, ""},
		{"max negative", "hello", -1, ""},
		{"empty string", "", 10, ""},
		{"fits entirely", "hello", 10, "hello"},
		{"truncate ASCII", longASCII, 5, "aaaaa"},
		{"exact fit", "abcde", 5, "abcde"},
		{"emoji two units max one yields empty", emoji, 1, ""},
		{"emoji fits when max two", emoji, 2, emoji},
		{"single ASCII before emoji when max one", "x" + emoji, 1, "x"},
		{"truncate before emoji when emoji would not fit", "ab" + emoji, 3, "ab"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TruncateStringForDiscord(tt.s, tt.maxUTF16)
			if got != tt.want {
				t.Errorf("TruncateStringForDiscord(%q, %d) = %q, want %q", tt.s, tt.maxUTF16, got, tt.want)
			}
			if UTF16Len(got) > tt.maxUTF16 && tt.maxUTF16 > 0 {
				t.Errorf("result longer than max: UTF16Len(%q) = %d > %d", got, UTF16Len(got), tt.maxUTF16)
			}
		})
	}
}

func TestTruncateStringForDiscord_ResultNeverExceedsMax(t *testing.T) {
	s := "a\U0001f600b\U0001f917c" // mix of ASCII and emoji
	for max := 1; max <= 20; max++ {
		out := TruncateStringForDiscord(s, max)
		if max > 0 && UTF16Len(out) > max {
			t.Fatalf("max=%d: UTF16Len(%q)=%d", max, out, UTF16Len(out))
		}
	}
}
