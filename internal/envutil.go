package internal

import "os"

// ============================================================================
// env_utils
// ============================================================================

func LookupEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
