package mask

// Mask returns a masked representation of value, counting characters by rune
// so that multi-byte characters count as one:
//
//   - empty string: returns ""
//   - fewer than 4 runes: returns "***"
//   - 4 to 7 runes: returns the first rune followed by "***"
//   - 8 or more runes: returns the first two runes, "***", and the last two
//
// Mask never returns the original value for any non-empty input.
func Mask(value string) string {
	r := []rune(value)
	switch n := len(r); {
	case n == 0:
		return ""
	case n < 4:
		return "***"
	case n < 8:
		return string(r[0]) + "***"
	default:
		return string(r[:2]) + "***" + string(r[n-2:])
	}
}
