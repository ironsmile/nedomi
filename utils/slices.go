package utils

// CopyStringSlice deeply copies a slice of strings
func CopyStringSlice(from []string) []string {
	res := make([]string, len(from))
	copy(res, from)
	return res
}
