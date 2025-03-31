package database

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// DB represents a database connection
type DB struct {
	*sql.DB
}

var db *DB

// InitDB initializes the database connection
func InitDB() (*DB, error) {
	if db != nil {
		return db, nil
	}

	// For now, we'll use SQLite
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "smsgw.db"
	}

	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	// Check connection
	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	// Create tables if they don't exist
	if err := createTables(conn); err != nil {
		return nil, fmt.Errorf("failed to create tables: %v", err)
	}

	db = &DB{conn}
	return db, nil
}

// createTables creates the necessary tables if they don't exist
func createTables(db *sql.DB) error {
	// Create subscriptions table
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS subscriptions (
		id INTEGER PRIMARY KEY,
		subscription_id TEXT NOT NULL,
		email TEXT NOT NULL,
		status TEXT NOT NULL,
		expiry_date TIMESTAMP,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_subscriptions_email ON subscriptions(email);
	CREATE INDEX IF NOT EXISTS idx_subscriptions_subscription_id ON subscriptions(subscription_id);
	`)
	if err != nil {
		return err
	}

	// Create phone_subscriptions table
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS phone_subscriptions (
		id INTEGER PRIMARY KEY,
		phone_number TEXT NOT NULL,
		subscription_id INTEGER,
		rate_limit INTEGER DEFAULT 5,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (subscription_id) REFERENCES subscriptions(id)
	);
	CREATE UNIQUE INDEX IF NOT EXISTS idx_phone_subscriptions_phone ON phone_subscriptions(phone_number);
	`)
	if err != nil {
		return err
	}

	// Create message_rate_limit table (replacing message_log with privacy-focused approach)
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS message_rate_limit (
		id INTEGER PRIMARY KEY,
		phone_number TEXT NOT NULL,
		sent_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		expiry TIMESTAMP DEFAULT (datetime('now', '+24 hours'))
	);
	CREATE INDEX IF NOT EXISTS idx_message_rate_limit_phone ON message_rate_limit(phone_number);
	CREATE INDEX IF NOT EXISTS idx_message_rate_limit_sent_at ON message_rate_limit(sent_at);
	CREATE INDEX IF NOT EXISTS idx_message_rate_limit_expiry ON message_rate_limit(expiry);
	`)
	if err != nil {
		return err
	}

	return nil
}

// SaveSubscription saves a new subscription
func (db *DB) SaveSubscription(subscriptionID, email, status string, expiryDate time.Time) error {
	_, err := db.Exec(`
	INSERT INTO subscriptions (subscription_id, email, status, expiry_date)
	VALUES (?, ?, ?, ?)
	`, subscriptionID, email, status, expiryDate)
	return err
}

// UpdateSubscriptionStatus updates a subscription's status
func (db *DB) UpdateSubscriptionStatus(subscriptionID, status string) error {
	_, err := db.Exec(`
	UPDATE subscriptions
	SET status = ?, updated_at = CURRENT_TIMESTAMP
	WHERE subscription_id = ?
	`, status, subscriptionID)
	return err
}

// LinkPhoneToSubscription links a phone number to a subscription
func (db *DB) LinkPhoneToSubscription(phoneNumber, email string) error {
	// First, get the subscription ID
	var subscriptionID int
	err := db.QueryRow(`
	SELECT id FROM subscriptions
	WHERE email = ? AND status = 'active'
	ORDER BY expiry_date DESC
	LIMIT 1
	`, email).Scan(&subscriptionID)
	if err != nil {
		return err
	}

	// Then, link the phone number
	_, err = db.Exec(`
	INSERT INTO phone_subscriptions (phone_number, subscription_id)
	VALUES (?, ?)
	ON CONFLICT(phone_number) DO UPDATE
	SET subscription_id = ?, updated_at = CURRENT_TIMESTAMP
	`, phoneNumber, subscriptionID, subscriptionID)
	return err
}

// UpdateRateLimitForPhone updates the rate limit for a phone number
func (db *DB) UpdateRateLimitForPhone(phoneNumber string, limit int) error {
	_, err := db.Exec(`
	INSERT INTO phone_subscriptions (phone_number, rate_limit)
	VALUES (?, ?)
	ON CONFLICT(phone_number) DO UPDATE
	SET rate_limit = ?, updated_at = CURRENT_TIMESTAMP
	`, phoneNumber, limit, limit)
	return err
}

// CheckRateLimit checks if a phone number has reached its rate limit
func (db *DB) CheckRateLimit(phoneNumber string) (bool, error) {
	// First, clean up expired entries
	if err := db.cleanupExpiredRateLimitEntries(); err != nil {
		return false, fmt.Errorf("failed to clean up expired entries: %v", err)
	}

	// Get the rate limit for this phone number
	var rateLimit int
	err := db.QueryRow(`
	SELECT COALESCE(rate_limit, 5) FROM phone_subscriptions
	WHERE phone_number = ?
	`, phoneNumber).Scan(&rateLimit)
	if err == sql.ErrNoRows {
		// Phone number not in database, use default limit of 5
		rateLimit = 5
	} else if err != nil {
		return false, err
	}

	// Count messages sent in the last hour
	var count int
	err = db.QueryRow(`
	SELECT COUNT(*) FROM message_rate_limit
	WHERE phone_number = ? AND sent_at > datetime('now', '-1 hour')
	`, phoneNumber).Scan(&count)
	if err != nil {
		return false, err
	}

	// Check if we're under the limit
	if count < rateLimit {
		// Log this message attempt for rate limiting purposes only
		_, err = db.Exec(`
		INSERT INTO message_rate_limit (phone_number)
		VALUES (?)
		`, phoneNumber)
		if err != nil {
			return false, err
		}
		return true, nil
	}

	return false, nil
}

// cleanupExpiredRateLimitEntries removes entries that have passed their expiry time
func (db *DB) cleanupExpiredRateLimitEntries() error {
	_, err := db.Exec(`DELETE FROM message_rate_limit WHERE expiry < datetime('now')`)
	return err
}

// PurgeOldMessageData purges message rate limit data older than the specified retention period
func (db *DB) PurgeOldMessageData() error {
	// Delete any entries older than 24 hours (failsafe in case expiry index isn't working)
	_, err := db.Exec(`DELETE FROM message_rate_limit WHERE sent_at < datetime('now', '-24 hours')`)
	return err
}
