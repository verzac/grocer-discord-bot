package utils

import "unicode/utf8"

// DiscordCheckboxOptionLabelMaxBytes is the UTF-8 byte limit for checkbox group option labels
// (enforced by the API in practice; keep labels ≤ 100 UTF-8 bytes).
const DiscordCheckboxOptionLabelMaxBytes = 100

// TruncateStringToMaxUTF8Bytes returns a prefix of s whose UTF-8 encoding is at most maxBytes long.
func TruncateStringToMaxUTF8Bytes(s string, maxBytes int) string {
	if maxBytes <= 0 {
		return ""
	}
	i := 0
	for i < len(s) {
		_, size := utf8.DecodeRuneInString(s[i:])
		if size == 0 {
			break
		}
		if i+size > maxBytes {
			break
		}
		i += size
	}
	return s[:i]
}
