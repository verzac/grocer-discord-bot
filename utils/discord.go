package utils

// DiscordCheckboxOptionLabelMaxUTF16 is the max label length for checkbox group options
// (Discord counts string length like JavaScript: UTF-16 code units; labels must be 1–100).
const DiscordCheckboxOptionLabelMaxUTF16 = 100

// UTF16UnitsForRune returns how many UTF-16 code units a single rune occupies (1 or 2).
func UTF16UnitsForRune(r rune) int {
	if r > 0xffff {
		return 2
	}
	return 1
}

// UTF16Len returns the length of s in UTF-16 code units, matching Discord/JS string length.
func UTF16Len(s string) int {
	n := 0
	for _, r := range s {
		n += UTF16UnitsForRune(r)
	}
	return n
}

// TruncateStringForDiscord returns a prefix of s whose UTF-16 length is at most maxUTF16.
func TruncateStringForDiscord(s string, maxUTF16 int) string {
	if maxUTF16 <= 0 {
		return ""
	}
	rs := []rune(s)
	n := 0
	i := 0
	for i < len(rs) {
		add := UTF16UnitsForRune(rs[i])
		if n+add > maxUTF16 {
			break
		}
		n += add
		i++
	}
	return string(rs[:i])
}
