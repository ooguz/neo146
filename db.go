package main

import (
	"database/sql"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func initDB() error {
	var err error
	db, err = sql.Open("sqlite3", "smsgw.db")
	if err != nil {
		return err
	}

	// Create rate limit table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS rate_limits (
			phone_number TEXT PRIMARY KEY,
			message_count INTEGER DEFAULT 0,
			last_reset TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			hourly_limit INTEGER DEFAULT 5
		)
	`)
	if err != nil {
		return err
	}

	// Create subscriptions table if it doesn't exist
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS subscriptions (
			id TEXT PRIMARY KEY,
			email TEXT,
			status TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			expires_at TIMESTAMP
		)
	`)
	if err != nil {
		return err
	}

	// Add phone_number column if it doesn't exist
	_, err = db.Exec(`
		ALTER TABLE subscriptions ADD COLUMN phone_number TEXT
	`)
	if err != nil && !strings.Contains(err.Error(), "duplicate column name") {
		return err
	}

	// Add hourly_limit column if it doesn't exist
	_, err = db.Exec(`
		ALTER TABLE subscriptions ADD COLUMN hourly_limit INTEGER DEFAULT 20
	`)
	if err != nil && !strings.Contains(err.Error(), "duplicate column name") {
		return err
	}

	return nil
}

func checkRateLimit(phoneNumber string) (bool, error) {
	var count int
	var lastReset time.Time
	var hourlyLimit int
	err := db.QueryRow(`
		SELECT message_count, last_reset, hourly_limit 
		FROM rate_limits 
		WHERE phone_number = ?
	`, phoneNumber).Scan(&count, &lastReset, &hourlyLimit)

	if err == sql.ErrNoRows {
		// First time user, create record
		_, err = db.Exec(`
			INSERT INTO rate_limits (phone_number, message_count, last_reset, hourly_limit)
			VALUES (?, 0, CURRENT_TIMESTAMP, 5)
		`, phoneNumber)
		return true, err
	}

	if err != nil {
		return false, err
	}

	// Check if we need to reset the counter (1 hour has passed)
	if time.Since(lastReset) > time.Hour {
		_, err = db.Exec(`
			UPDATE rate_limits 
			SET message_count = 0, last_reset = CURRENT_TIMESTAMP
			WHERE phone_number = ?
		`, phoneNumber)
		return true, err
	}

	// Check if user has reached the limit
	if count >= hourlyLimit {
		return false, nil
	}

	// Increment counter
	_, err = db.Exec(`
		UPDATE rate_limits 
		SET message_count = message_count + 1
		WHERE phone_number = ?
	`, phoneNumber)
	return true, err
}

func saveSubscription(subscriptionID, email, status string, expiresAt time.Time) error {
	_, err := db.Exec(`
		INSERT OR REPLACE INTO subscriptions (id, email, status, expires_at)
		VALUES (?, ?, ?, ?)
	`, subscriptionID, email, status, expiresAt)
	return err
}

func getSubscriptionStatus(subscriptionID string) (string, error) {
	var status string
	err := db.QueryRow(`
		SELECT status FROM subscriptions WHERE id = ?
	`, subscriptionID).Scan(&status)
	return status, err
}

func linkPhoneToSubscription(phoneNumber, email string) error {
	_, err := db.Exec(`
		UPDATE subscriptions 
		SET phone_number = ?, hourly_limit = 20
		WHERE email = ? AND status = 'active'
	`, phoneNumber, email)
	return err
}

func updateSubscriptionStatus(subscriptionID, status string) error {
	// First get the phone number associated with this subscription
	var phoneNumber sql.NullString
	err := db.QueryRow(`
		SELECT phone_number FROM subscriptions WHERE id = ?
	`, subscriptionID).Scan(&phoneNumber)
	if err != nil {
		return err
	}

	// Update subscription status
	_, err = db.Exec(`
		UPDATE subscriptions 
		SET status = ?
		WHERE id = ?
	`, status, subscriptionID)
	if err != nil {
		return err
	}

	// Only update rate limit if we have a phone number
	if phoneNumber.Valid {
		// Update rate limit based on subscription status
		hourlyLimit := 5 // Default limit for inactive/cancelled subscriptions
		if status == "active" {
			hourlyLimit = 20 // Higher limit for active subscriptions
		}

		// Update rate limit for the associated phone number
		_, err = db.Exec(`
			UPDATE rate_limits 
			SET hourly_limit = ?
			WHERE phone_number = ?
		`, hourlyLimit, phoneNumber.String)
		return err
	}

	return nil
}

func updateRateLimitForPhone(phoneNumber string, hourlyLimit int) error {
	_, err := db.Exec(`
		UPDATE rate_limits 
		SET hourly_limit = ?
		WHERE phone_number = ?
	`, hourlyLimit, phoneNumber)
	return err
}
