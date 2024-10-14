package forum

import (
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func HandleLogin(w http.ResponseWriter, r *http.Request) {
	if isAuthenticated(r, w) {
		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return
	}

	if r.Method == "POST" {
		username := r.FormValue("username")
		username = strings.ToUpper(username[:1]) + strings.ToLower(username[1:])

		password := r.FormValue("password")

		// Start a transaction
		tx, err := Db.Begin()
		if err != nil {
			log.Printf("Error starting transaction: %v", err)
			ErrorHandler(w, r, http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		// Find the user in the database
		var user User
		err = tx.QueryRow("SELECT user_id, username, email, password, date_created FROM users WHERE username = ?", username).Scan(&user.UserID, &user.Username, &user.Email, &user.Password, &user.DateCreated)
		if err != nil {
			log.Printf("Error querying user: %v", err)
			ErrorHandler(w, r, http.StatusInternalServerError)
			return
		}

		// Compare the hashed password
		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
		if err != nil {
			log.Printf("Password mismatch for username: %s", username)
			renderLoginPage(w, r, "Invalid username or password")
			return
		}
		// Create a new session
		sessionID := uuid.New().String()
		expiresAt := time.Now().Add(24 * time.Hour) // Set session to expire after 24 hours

		// Insert the session into the database
		_, err = Db.Exec("INSERT INTO sessions (id, user_id, created_at, expires_at) VALUES (?, ?, ?, ?)", sessionID, user.UserID, time.Now(), expiresAt)
		if err != nil {
			log.Printf("Error inserting session: %v", err)
			ErrorHandler(w, r, http.StatusInternalServerError)
			return
		}

		// Update user's online status to true
		_, err = Db.Exec("UPDATE users SET is_online = TRUE WHERE user_id = ?", user.UserID)
		if err != nil {
			log.Printf("Error updating online status: %v", err)
			fmt.Printf("%d is online", user.UserID)
			ErrorHandler(w, r, http.StatusInternalServerError)
			return
		}

		// Update the user's last active time
		err = updateUserLastActivity(user.UserID)
		if err != nil {
			log.Printf("Error updating last active time: %v", err)
			ErrorHandler(w, r, http.StatusInternalServerError)
			return
		}

		// Commit the transaction
		err = tx.Commit()
		if err != nil {
			log.Printf("Error committing transaction: %v", err)
			ErrorHandler(w, r, http.StatusInternalServerError)
			return
		}

		// Set the session cookie
		cookie := &http.Cookie{
			Name:     "forum_session",
			Value:    sessionID,
			Expires:  expiresAt,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode, // Change to http.SameSiteNoneMode for testing
		}
		http.SetCookie(w, cookie)
		log.Printf("Set-Cookie: %s=%s; Path=%s; Expires=%s; HttpOnly=%t; SameSite=%s",
			cookie.Name, cookie.Value, cookie.Path, cookie.Expires, cookie.HttpOnly, cookie.SameSite)

		log.Printf("Login successful for user: %s, session ID: %s", username, sessionID)
		// Redirect to the home page
		http.Redirect(w, r, "/home", http.StatusSeeOther) // ... (rest of the login process remains the same)

		validateSession(r, w, sessionID)
	} else {
		renderLoginPage(w, r, "")
	}
}

func renderLoginPage(w http.ResponseWriter, r *http.Request, errorMessage string) {
	t, err := template.ParseFiles("templates/login.html")
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}
	data := struct {
		ErrorMessage string
	}{
		ErrorMessage: errorMessage,
	}
	err = t.Execute(w, data)
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}
}

func HandleLogout(w http.ResponseWriter, r *http.Request) {
	// Get the session ID from the cookie
	sessionID, _ := getCookie(r, w, CookieName)
	var userID int
	err := Db.QueryRow("SELECT user_id FROM sessions WHERE id = ?", sessionID).Scan(&userID)
	if err != nil {
		log.Println("Error retrieving user ID:", err)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get the session cookie
	cookie, err := r.Cookie("forum_session")
	if err != nil {
		if err == http.ErrNoCookie {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}

	// Start a transaction
	tx, err := Db.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}
	defer func() {
		if err != nil {
			if rErr := tx.Rollback(); rErr != nil {
				log.Printf("Error rolling back transaction: %v", rErr)
			}
		}
	}()

	// Delete the session from the database
	_, err = tx.Exec("DELETE FROM sessions WHERE id = ?", cookie.Value)
	if err != nil {
		log.Printf("Error deleting session: %v", err)
		return // This will trigger the rollback
	}

	// Update user's online status to false
	_, err = tx.Exec("UPDATE users SET is_online = FALSE WHERE user_id = ?", userID)
	if err != nil {
		log.Printf("Error updating online status: %v", err)
		return // This will trigger the rollback
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		log.Printf("Error committing transaction: %v", err)
		ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}

	// Invalidate the session cookie
	cookie = &http.Cookie{
		Name:     "forum_session",
		Value:    "",
		Path:     "/",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)

	// Redirect to login page
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func UpdateLastActive(w http.ResponseWriter, r *http.Request) error {
	// Get the session ID from the cookie
	sessionID, cookieErr := getCookie(r, w, CookieName)
	if !cookieErr {
		return errors.New("failed to get cookie")
	}

	// Get the user ID from the session
	var userID int
	queryErr := Db.QueryRow("SELECT user_id FROM sessions WHERE id = ?", sessionID).Scan(&userID)
	if queryErr != nil {
		if queryErr == sql.ErrNoRows {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return nil
		}
		return queryErr
	}

	// Update the last_active timestamp for the user
	now := time.Now()
	result, execErr := Db.Exec("UPDATE users SET last_active = ? WHERE user_id = ?", now, userID)
	if execErr != nil {
		return execErr
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("no rows were updated")
	}

	// Verify the update
	var lastActive time.Time
	verifyErr := Db.QueryRow("SELECT last_active FROM users WHERE user_id = ?", userID).Scan(&lastActive)
	if verifyErr != nil {
		return verifyErr
	}

	log.Printf("Updated last_active to %v for user %d", lastActive, userID)

	return nil
}

// Last activity

func updateUserLastActivity(userID int) error {
	now := time.Now()

	// Start a transaction
	tx, err := Db.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		return err
	}
	defer tx.Rollback()

	// Update the last_active timestamp
	_, err = tx.Exec("UPDATE users SET last_active = ? WHERE user_id = ?", now, userID)
	if err != nil {
		log.Printf("Error updating last_active: %v", err)
		return err
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		log.Printf("Error committing transaction: %v", err)
		return err
	}

	return nil
}
