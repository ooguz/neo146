package utils

import (
	"encoding/base64"
	"fmt"
	"regexp"
)

// IsURL checks if the text is a URL
func IsURL(text string) bool {
	re := regexp.MustCompile(`https?://[\w\-\.]+[\w\-/]*`)
	return re.MatchString(text)
}

// SplitMessage splits a message into parts of maximum length
func SplitMessage(message string, maxLength int) []string {
	// Handle empty message case
	if message == "" {
		return []string{}
	}

	// Convert to runes to handle Unicode characters properly
	messageRunes := []rune(message)

	// If the message fits in one chunk, return it
	if len(messageRunes) <= maxLength {
		return []string{message}
	}

	var parts []string

	// Process the message in chunks
	for i := 0; i < len(messageRunes); i += maxLength {
		end := i + maxLength
		if end > len(messageRunes) {
			end = len(messageRunes)
		}

		parts = append(parts, string(messageRunes[i:end]))
	}

	return parts
}

// SplitAndEncodeMessage splits a message into parts and encodes each part
func SplitAndEncodeMessage(message string, maxLength int) []string {
	// Handle empty message case
	if message == "" {
		return []string{}
	}

	// Split the message into parts
	parts := SplitMessage(message, maxLength)

	// Encode each part with a header
	var encodedParts []string
	for i, part := range parts {
		// First encode the message part
		encoded := base64.StdEncoding.EncodeToString([]byte(part))
		// Then add the header to the encoded content
		header := fmt.Sprintf("GW%d|", i+1)
		encodedParts = append(encodedParts, header+encoded)
	}

	return encodedParts
}

// Min returns the minimum of two integers
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
