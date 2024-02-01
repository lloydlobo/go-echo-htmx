package internal

// ============================================================================
// string_utils
// ============================================================================

func Contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}

	return false
}
