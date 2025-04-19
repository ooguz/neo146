package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

var (
	lockFile = filepath.Join(os.TempDir(), "neo146_telegram_bot.lock")
)

// AcquireLock attempts to acquire a file-based lock
func AcquireLock() error {
	// Check if lock file exists
	if _, err := os.Stat(lockFile); err == nil {
		// Lock file exists, check if it's stale (older than 5 minutes)
		info, err := os.Stat(lockFile)
		if err != nil {
			return fmt.Errorf("error checking lock file: %v", err)
		}
		if time.Since(info.ModTime()) < 5*time.Minute {
			return fmt.Errorf("bot is already running")
		}
		// Remove stale lock file
		os.Remove(lockFile)
	}

	// Create lock file
	file, err := os.Create(lockFile)
	if err != nil {
		return fmt.Errorf("error creating lock file: %v", err)
	}
	file.Close()

	return nil
}

// ReleaseLock releases the file-based lock
func ReleaseLock() error {
	return os.Remove(lockFile)
}
