package utils

import (
	"testing"
)

func TestIsURL(t *testing.T) {
	validURLs := []string{
		"http://example.com",
		"https://example.com",
		"http://www.example.com/path/to/resource",
		"https://example.com/path?query=value",
		"https://subdomain.example.co.uk/path",
	}

	invalidURLs := []string{
		"",
		"example.com",
		"www.example.com",
		"http://",
		"https://",
		"just some text",
		"http:/example.com", // Missing a slash
	}

	// Test valid URLs
	for _, url := range validURLs {
		if !IsURL(url) {
			t.Errorf("Expected %s to be recognized as a valid URL, but it wasn't", url)
		}
	}

	// Test invalid URLs
	for _, url := range invalidURLs {
		if IsURL(url) {
			t.Errorf("Expected %s to be recognized as an invalid URL, but it was considered valid", url)
		}
	}
}

func TestSplitMessage(t *testing.T) {
	testCases := []struct {
		name          string
		input         string
		maxChunkSize  int
		expectedParts int
	}{
		{
			name:          "Empty message",
			input:         "",
			maxChunkSize:  100,
			expectedParts: 0,
		},
		{
			name:          "Short message, no splitting needed",
			input:         "This is a short message",
			maxChunkSize:  100,
			expectedParts: 1,
		},
		{
			name:          "Message exactly at max size",
			input:         "12345678901234567890", // 20 characters
			maxChunkSize:  20,
			expectedParts: 1,
		},
		{
			name:          "Message needs splitting",
			input:         "This is a longer message that needs to be split into multiple parts",
			maxChunkSize:  20,
			expectedParts: 4,
		},
		{
			name:          "Message with emoji and special characters",
			input:         "Hello 👋 world! Special chars: äöü",
			maxChunkSize:  10,
			expectedParts: 4,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := SplitMessage(tc.input, tc.maxChunkSize)

			if len(result) != tc.expectedParts {
				t.Errorf("Expected %d parts, got %d", tc.expectedParts, len(result))
			}

			// Check that no part exceeds the max chunk size
			for i, part := range result {
				if len([]rune(part)) > tc.maxChunkSize {
					t.Errorf("Part %d exceeds max chunk size: %d > %d",
						i, len([]rune(part)), tc.maxChunkSize)
				}
			}

			// Reconstruct the message and check it matches the input
			if tc.input != "" {
				reconstructed := ""
				for _, part := range result {
					reconstructed += part
				}
				if reconstructed != tc.input {
					t.Errorf("Reconstructed message doesn't match input. Expected: %s, Got: %s",
						tc.input, reconstructed)
				}
			}
		})
	}
}

func TestSplitAndEncodeMessage(t *testing.T) {
	testCases := []struct {
		name          string
		input         string
		maxChunkSize  int
		expectedParts int
	}{
		{
			name:          "Empty message",
			input:         "",
			maxChunkSize:  100,
			expectedParts: 0,
		},
		{
			name:          "Short message, no splitting needed",
			input:         "This is a short message",
			maxChunkSize:  100,
			expectedParts: 1,
		},
		{
			name:          "Message needs splitting",
			input:         "This is a longer message that needs to be split into multiple parts and then encoded for transmission via SMS",
			maxChunkSize:  30,
			expectedParts: 4,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := SplitAndEncodeMessage(tc.input, tc.maxChunkSize)

			if len(result) != tc.expectedParts {
				t.Errorf("Expected %d parts, got %d", tc.expectedParts, len(result))
			}

			// Since encoding could change the length, we'll just check that parts exist
			if tc.expectedParts > 0 && len(result) == 0 {
				t.Errorf("Expected at least one part, got none")
			}
		})
	}
}

func TestMin(t *testing.T) {
	testCases := []struct {
		a, b, expected int
	}{
		{5, 10, 5},
		{10, 5, 5},
		{7, 7, 7},
		{-5, 5, -5},
		{0, 10, 0},
	}

	for _, tc := range testCases {
		result := Min(tc.a, tc.b)
		if result != tc.expected {
			t.Errorf("Min(%d, %d) = %d, expected %d", tc.a, tc.b, result, tc.expected)
		}
	}
}
