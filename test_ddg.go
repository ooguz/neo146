//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
	// Initialize HTTP client
	httpClient = &http.Client{
		Timeout: 10 * time.Second,
	}

	// Test the DuckDuckGo search
	query := "golang programming"
	fmt.Printf("Testing search for '%s'...\n", query)

	results, err := fetchDuckDuckGoResults(query)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("Search results:")
	fmt.Println("======================================")
	fmt.Println(results)
	fmt.Println("======================================")
}
