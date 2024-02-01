package internal

import (
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
