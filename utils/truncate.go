package utils

const (
	defaultEllipticSuffix = "..."
)

// TruncateStringWithTargetLength ...
func TruncateStringWithTargetLength(str string, targetLength int) string {
	if targetLength <= 0 {
		return ""
	}
	targetLength -= 3

	// This code cannot support Japanese
	// orgLen := len(str)
	// if orgLen <= length {
	//     return str
	// }
	// return str[:length]

	// Support Japanese
	// Ref: Range loops https://blog.golang.org/strings
	truncated := ""
	count := 0
	doesExceed := false
	for _, char := range str {
		truncated += string(char)
		count++
		if count >= targetLength {
			doesExceed = true
			break
		}
	}
	if doesExceed {
		return truncated + defaultEllipticSuffix
	}
	return truncated
}
