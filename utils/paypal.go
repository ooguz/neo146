package utils

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

// ParseFloat parses a string to a float64
func ParseFloat(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

// ParsePayPalDate parses a PayPal date string to time.Time
func ParsePayPalDate(dateStr string) time.Time {
	// PayPal dates are in format: HH:MM:SS MMM DD, YYYY PST
	// Example: 15:30:45 Jan 18, 2009 PST
	loc, _ := time.LoadLocation("America/Los_Angeles")
	t, _ := time.ParseInLocation("15:04:05 Jan 02, 2006 MST", dateStr, loc)
	return t
}

// VerifyPayPalIPN verifies a PayPal IPN message
func VerifyPayPalIPN(body []byte) error {
	// Create a new request to PayPal
	req, err := http.NewRequest("POST", "https://ipnpb.paypal.com/cgi-bin/webscr", bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	// Add the cmd=_notify-validate parameter
	q := req.URL.Query()
	q.Add("cmd", "_notify-validate")
	req.URL.RawQuery = q.Encode()

	// Set the content type
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Read the response
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Check if the response is "VERIFIED"
	if string(responseBody) != "VERIFIED" {
		return fmt.Errorf("IPN not verified by PayPal")
	}

	return nil
}
