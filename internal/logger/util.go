package logger

// Truncate returns the string truncated to maxLen runes. If the string is longer, it appends "...".
func Truncate(s string, maxLen int) string {
	r := []rune(s)
	if len(r) <= maxLen {
		return s
	}
	return string(r[:maxLen]) + "..."
}
