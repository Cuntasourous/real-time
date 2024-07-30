package forum
import (
    "database/sql"
    "log"
    "net/http"
    "time"
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
func getCookie(r *http.Request, w http.ResponseWriter, name string) (string, bool) {
    cookie, err := r.Cookie(name)
    if err != nil {
        return "", false
    }
    return cookie.Value, true
}
func isAuthenticated(r *http.Request, w http.ResponseWriter) bool {
    cookie, err := r.Cookie("forum_session")
    if err != nil {
        log.Println("No session cookie found")
        log.Printf("Cookies received: %v", r.Cookies())
        return false
    }
    log.Printf("Session cookie found: %s", cookie.Value)
    // Validate the session ID from the cookie with your session store
    return validateSession(r, w, cookie.Value)
}
// validateSession checks if the session ID exists and is still valid
func validateSession(r *http.Request, w http.ResponseWriter, sessionID string) bool {
    var expiresAt time.Time
    var userID int
    // Query the database for the session
    err := Db.QueryRow("SELECT expires_at, user_id FROM sessions WHERE id = ?", sessionID).Scan(&expiresAt, &userID)
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
    // Count the number of sessions for the user
    var count int
    err = Db.QueryRow("SELECT COUNT(*) FROM sessions WHERE user_id = ?", userID).Scan(&count)
    if err != nil {
        log.Printf("Error counting number of sessions: %v", err)
        return false
    }
    // Delete the oldest session if more than one session exists
    if count > 1 {
        var oldestSessionID string
        err = Db.QueryRow("SELECT id FROM sessions WHERE user_id = ? ORDER BY created_at ASC LIMIT 1", userID).Scan(&oldestSessionID)
        if err != nil {
            log.Printf("Error fetching oldest session ID: %v", err)
            return false
        }
        _, err = Db.Exec("DELETE FROM sessions WHERE id = ?", oldestSessionID)
        if err != nil {
            log.Printf("Error deleting oldest session: %v", err)
            return false
        }
        log.Printf("Deleted oldest session: %s", oldestSessionID)
        // Update the expiration time of the current session
        newExpiration := time.Now().Add(4 * time.Hour) // Example: extending the session by 4 hours
        _, err = Db.Exec("UPDATE sessions SET expires_at = ? WHERE id = ?", newExpiration, sessionID)
        if err != nil {
            log.Printf("Error updating session expiration: %v", err)
            return false
        }
        log.Printf("Updated expiration time for session: %s", sessionID)
    }
    return true
}