package forum

import (
	"database/sql"
	"log"
	"time"
)

// validateSession checks if the session ID exists and is still valid
func validateSession(sessionID string) bool {
	var expiresAt time.Time

	// Query the database for the session
	err := Db.QueryRow("SELECT expires_at FROM sessions WHERE id = ?", sessionID).Scan(&expiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Session ID not found: %s", sessionID)
			return false
		}
		log.Printf("Error querying session: %v", err)
		return false
	}

	// Check if the session has expired
	if time.Now().After(expiresAt) {
		log.Printf("Session ID expired: %s", sessionID)
		return false
	}

	return true
}
