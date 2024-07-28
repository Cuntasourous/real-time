package forum

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// func HandleLogin(w http.ResponseWriter, r *http.Request) {
// 	if isAuthenticated(r) {
// 		http.Redirect(w, r, "/home", http.StatusSeeOther)
// 		return
// 	}

// 	if r.Method == "POST" {
// 		username := r.FormValue("username")
// 		password := r.FormValue("password")

// 		// Basic validation
// 		if username == "" || password == "" {
// 			// http.Error(w, "Please fill in all fields", http.StatusBadRequest)
// 			ErrorHandler(w, r, http.StatusBadRequest)
// 			return
// 		}

// 		// Start a transaction
// 		tx, err := Db.Begin()
// 		if err != nil {
// 			log.Printf("Error starting transaction: %v", err)
// 			ErrorHandler(w, r, http.StatusInternalServerError)
// 			return
// 		}
// 		defer func() {
// 			if rErr := tx.Rollback(); rErr != nil && err == nil {
// 				log.Printf("Error rolling back transaction: %v", rErr)
// 			}
// 		}()

// 		// Find the user in the database
// 		var user User
// 		err = tx.QueryRow("SELECT user_id, username, email, password, date_created FROM users WHERE username = ?", username).Scan(&user.UserID, &user.Username, &user.Email, &user.Password, &user.DateCreated)
// 		if err != nil {
// 			if err == sql.ErrNoRows {
// 				log.Printf("User not found for username: %s", username)
// 				// http.Error(w, "Invalid username or password", http.StatusBadRequest)
// 				ErrorHandler(w, r, http.StatusBadRequest)
// 				//instead of a bad request, we can add an error message
// 				return
// 			}
// 			log.Printf("Error querying user: %v", err)
// 			ErrorHandler(w, r, http.StatusInternalServerError)
// 			return
// 		}

// 		// Compare the hashed password
// 		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
// 		if err != nil {
// 			log.Printf("Password mismatch for username: %s", username)
// 			// http.Error(w, "Invalid username or password", http.StatusBadRequest)
// 			ErrorHandler(w, r, http.StatusBadRequest)
// 			return
// 		}

// 		// Create a new session
// 		sessionID := uuid.New().String()
// 		expiresAt := time.Now().Add(24 * time.Hour) // Set session to expire after 24 hours

// 		// Insert the session into the database
// 		_, err = Db.Exec("INSERT INTO sessions (id, user_id, created_at, expires_at) VALUES (?, ?, ?, ?)", sessionID, user.UserID, time.Now(), expiresAt)
// 		if err != nil {
// 			log.Printf("Error inserting session: %v", err)
// 			ErrorHandler(w, r, http.StatusInternalServerError)
// 			return
// 		}

// 		// Commit the transaction
// 		err = tx.Commit()
// 		if err != nil {
// 			log.Printf("Error committing transaction: %v", err)
// 			ErrorHandler(w, r, http.StatusInternalServerError)
// 			return
// 		}

// 		// Set the session cookie
// 		cookie := &http.Cookie{
// 			Name:     "forum_session",
// 			Value:    sessionID,
// 			Expires:  expiresAt,
// 			Path:     "/",
// 			HttpOnly: true,
// 			SameSite: http.SameSiteLaxMode, // Change to http.SameSiteNoneMode for testing
// 		}
// 		http.SetCookie(w, cookie)
// 		log.Printf("Set-Cookie: %s=%s; Path=%s; Expires=%s; HttpOnly=%t; SameSite=%s",
// 			cookie.Name, cookie.Value, cookie.Path, cookie.Expires, cookie.HttpOnly, cookie.SameSite)

// 		log.Printf("Login successful for user: %s, session ID: %s", username, sessionID)
// 		// Redirect to the home page
// 		http.Redirect(w, r, "/home", http.StatusSeeOther)
// 	} else {
// 		t, err := template.ParseFiles("templates/login.html")
// 		if err != nil {
// 			ErrorHandler(w, r, http.StatusInternalServerError)
// 			return
// 		}
// 		err = t.Execute(w, nil)
// 		if err != nil {
// 			ErrorHandler(w, r, http.StatusInternalServerError)
// 			return
// 		}
// 	}
// }

func HandleLogin(w http.ResponseWriter, r *http.Request) {
	if isAuthenticated(r) {
		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return
	}

	if r.Method == "POST" {
		username := r.FormValue("username")
		password := r.FormValue("password")

		// Basic validation
		if username == "" || password == "" {
			renderLoginPage(w, r, "Please fill in all fields")
			return
		}

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
			if err == sql.ErrNoRows {
				log.Printf("User not found for username: %s", username)
				renderLoginPage(w, r, "Invalid username or password")
				return
			}
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

		validateSession(sessionID)
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
	// Get the session cookie
	cookie, err := r.Cookie("forum_session")
	if err != nil {
		if err == http.ErrNoCookie {
			// No session to log out from
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
		if rErr := tx.Rollback(); rErr != nil && err == nil {
			log.Printf("Error rolling back transaction: %v", rErr)
		}
	}()

	// Delete the session from the database
	_, err = tx.Exec("DELETE FROM sessions WHERE id = ?", cookie.Value)
	if err != nil {
		log.Printf("Error deleting session: %v", err)
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

	// Invalidate the session cookie
	cookie = &http.Cookie{
		Name:     "forum_session",
		Value:    "",
		Path:     "/",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode, // Change to http.SameSiteNoneMode for testing
	}
	http.SetCookie(w, cookie)

	// Redirect to login page
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
