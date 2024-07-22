package forum

import (
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
