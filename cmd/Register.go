package forum

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		username := r.FormValue("username")
		email := r.FormValue("email")
		password := r.FormValue("password")

		if username == "" || email == "" || password == "" {
			// http.Error(w, "Please fill in all fields", http.StatusBadRequest)
			ErrorHandler(w, r, http.StatusBadRequest)
			return
		}

		// Hash the password
		hashedPassword, err := hashPassword(password)
		if err != nil {
			log.Printf("Error hashing password: %v", err)
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
		defer tx.Rollback() // Roll back the transaction if it's not committed

		// Prepare the SQL statement
		stmt, err := tx.Prepare("INSERT INTO users(username, email, password) VALUES(?, ?, ?)")
		if err != nil {
			log.Printf("Error preparing SQL statement: %v", err)
			// http.Error(w, "Error preparing SQL statement", http.StatusInternalServerError)
			ErrorHandler(w, r, http.StatusInternalServerError)
			return
		}
		defer stmt.Close()

		// Execute the statement
		result, err := stmt.Exec(username, email, hashedPassword)
		if err != nil {
			log.Printf("Error inserting user: %v", err)
			// http.Error(w, "Error inserting user", http.StatusInternalServerError)
			ErrorHandler(w, r, http.StatusInternalServerError)
			return
		}

		// Commit the transaction
		err = tx.Commit()
		if err != nil {
			log.Printf("Error committing transaction: %v", err)
			// http.Error(w, "Error committing transaction", http.StatusInternalServerError)
			ErrorHandler(w, r, http.StatusInternalServerError)
			return
		}

		lastInsertID, err := result.LastInsertId()
		if err != nil {
			log.Printf("Error getting last inserted ID: %v", err)
		} else {
			log.Printf("Last inserted ID: %d", lastInsertID)
		}

		// Create a new session
		sessionID := uuid.New().String()
		expiresAt := time.Now().Add(24 * time.Hour) // Set session to expire after 24 hours

		rows, err := Db.Query("SELECT user_id FROM users WHERE username = ?", username)
		if err != nil {
			fmt.Println("error: ", err )
			ErrorHandler(w, r, http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var userID int
		for rows.Next() {
			if err := rows.Scan(&userID); err != nil {
				ErrorHandler(w, r, http.StatusInternalServerError)
				return
			}
		}

		// Insert the session into the database
		_, err = Db.Exec("INSERT INTO sessions (id, user_id, created_at, expires_at) VALUES (?, ?, ?, ?)", sessionID, userID, time.Now(), expiresAt)
		if err != nil {
			log.Printf("Error inserting session: %v", err)
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

		log.Printf("Sign uo successful for user: %s, session ID: %s", username, sessionID)

		http.Redirect(w, r, "/home", http.StatusSeeOther)
	} else {
		t, err := template.ParseFiles("templates/register.html")
		if err != nil {
			// http.Error(w, "Error parsing template", http.StatusInternalServerError)
			ErrorHandler(w, r, http.StatusInternalServerError)
			return
		}
		err = t.Execute(w, nil)
		if err != nil {
			ErrorHandler(w, r, http.StatusInternalServerError)
			return
		}
	}
}
