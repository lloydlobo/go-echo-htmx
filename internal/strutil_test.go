package internal

import (
	"encoding/base64"
	"testing"
)

func TestContains(t *testing.T) {
	tests := []struct {
		slice    []string
		str      string
		expected bool
	}{
		{[]string{"apple", "banana", "orange"}, "banana", true},
		{[]string{"apple", "banana", "orange"}, "grape", false},
		{[]string{"one", "two", "three"}, "two", true},
		{[]string{}, "empty", false},
	}

	for _, test := range tests {
		result := Contains(test.slice, test.str)
		if result != test.expected {
			t.Errorf("contains(%v, %s) = %t, expected %t", test.slice, test.str, result, test.expected)
		}
	}
}

func TestGenRandStr(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{"test1", 10},
		{"test2", 20},
		{"test3", 30},
		{"", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			randomStr, err := GenRandStr(tt.length)

			// Check for errors
			if err != nil {
				t.Errorf("Error generating random string: %v", err)
			}

			// Check the length of the generated string
			if len(randomStr) != base64.URLEncoding.EncodedLen(tt.length) {
				t.Errorf("Expected length %d, got length %d", tt.length, len(randomStr))
			}

			// Decode the string to check if it's a valid base64 encoding
			decodedBytes, decodeErr := base64.URLEncoding.DecodeString(randomStr)
			if decodeErr != nil {
				t.Errorf("Error decoding base64 string: %v", decodeErr)
			}

			// Check if the decoded length matches the input length
			if len(decodedBytes) != tt.length {
				t.Errorf("Expected decoded length %d, got length %d", tt.length, len(decodedBytes))
			}
		})
	}
}
