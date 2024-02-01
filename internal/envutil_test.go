package internal

import (
	"os"
	"testing"
)

func TestLookupEnv(t *testing.T) {
	// Test case 1: Key exists in the environment
	key1 := "EXISTING_KEY"
	defaultValue1 := "DEFAULT_VALUE_1"
	os.Setenv(key1, "EXISTING_VALUE_1")
	defer os.Unsetenv(key1)

	result1 := LookupEnv(key1, defaultValue1)
	if result1 != "EXISTING_VALUE_1" {
		t.Errorf("Expected %s, got %s", "EXISTING_VALUE_1", result1)
	}

	// Test case 2: Key does not exist in the environment
	key2 := "NON_EXISTING_KEY"
	defaultValue2 := "DEFAULT_VALUE_2"

	result2 := LookupEnv(key2, defaultValue2)
	if result2 != defaultValue2 {
		t.Errorf("Expected %s, got %s", defaultValue2, result2)
	}
}
