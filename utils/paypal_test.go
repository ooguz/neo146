package utils

import (
	"testing"
	"time"
)

func TestParseFloat(t *testing.T) {
	testCases := []struct {
		name           string
		input          string
		expectedOutput float64
	}{
		{
			name:           "Valid integer",
			input:          "123",
			expectedOutput: 123.0,
		},
		{
			name:           "Valid decimal",
			input:          "123.45",
			expectedOutput: 123.45,
		},
		{
			name:           "Invalid input",
			input:          "abc",
			expectedOutput: 0,
		},
		{
			name:           "Empty input",
			input:          "",
			expectedOutput: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ParseFloat(tc.input)
			if result != tc.expectedOutput {
				t.Errorf("Expected %f, got %f", tc.expectedOutput, result)
			}
		})
	}
}

func TestParsePayPalDate(t *testing.T) {
	testCases := []struct {
		name       string
		input      string
		expectZero bool
	}{
		{
			name:       "Valid date",
			input:      "21:12:52 Apr 10, 2024 PDT",
			expectZero: false,
		},
		{
			name:       "Invalid format",
			input:      "2024-04-10 21:12:52",
			expectZero: true,
		},
		{
			name:       "Empty string",
			input:      "",
			expectZero: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ParsePayPalDate(tc.input)

			if tc.expectZero && !result.IsZero() {
				t.Errorf("Expected zero time, got %v", result)
			}

			if !tc.expectZero {
				// Successful parsing should result in a non-zero time
				if result.IsZero() {
					t.Errorf("Expected non-zero time, got zero time")
				}

				// If the date is valid, check if month is April (4) for our test case
				if result.Month() != time.April {
					t.Errorf("Expected month to be April, got %v", result.Month())
				}
			}
		})
	}
}

// Note: VerifyPayPalIPN is difficult to test properly without mocking HTTP clients,
// which would require making significant changes to the function itself to accept
// a custom client for testing. In a real application, this function should be
// refactored to accept a custom HTTP client for better testability.
