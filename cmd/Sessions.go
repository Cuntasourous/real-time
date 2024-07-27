package forum

import (
	"database/sql"
	"log"
	"time"
	"net/http"
)



// Helper function to set a cookie
func SetCookie(w http.ResponseWriter, name string, value string, expires time.Time) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		Expires:  expires,
		Path:     "/",
		HttpOnly: true,
	}
	http.SetCookie(w, cookie)
}

// Helper function to get a cookie value
func getCookie(r *http.Request, name string) (string, bool) {
	cookie, err := r.Cookie(name)
	if err != nil {
		return "", false
	}
	return cookie.Value, true
}

func isAuthenticated(r *http.Request) bool {
	cookie, err := r.Cookie("forum_session")
	if err != nil {
		log.Println("No session cookie found")
		log.Printf("Cookies received: %v", r.Cookies())
		return false
	}

	log.Printf("Session cookie found: %s", cookie.Value)
	// Validate the session ID from the cookie with your session store
	return validateSession(cookie.Value)
}


// validateSession checks if the session ID exists and is still valid
func validateSession(sessionID string) bool {
	var expiresAt time.Time
	var 

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
