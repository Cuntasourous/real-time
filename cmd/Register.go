package forum

import (
	"html/template"
	"log"
	"net/http"
)

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		username := r.FormValue("username")
		email := r.FormValue("email")
		password := r.FormValue("password")

		if username == "" || email == "" || password == "" {
			http.Error(w, "Please fill in all fields", http.StatusBadRequest)
			return
		}

		// Hash the password
		hashedPassword, err := hashPassword(password)
		if err != nil {
			log.Printf("Error hashing password: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Start a transaction
		tx, err := Db.Begin()
		if err != nil {
			log.Printf("Error starting transaction: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		defer tx.Rollback() // Roll back the transaction if it's not committed

		// Prepare the SQL statement
		stmt, err := tx.Prepare("INSERT INTO users(username, email, password) VALUES(?, ?, ?)")
		if err != nil {
			log.Printf("Error preparing SQL statement: %v", err)
			http.Error(w, "Error preparing SQL statement", http.StatusInternalServerError)
			return
		}
		defer stmt.Close()

		// Execute the statement
		result, err := stmt.Exec(username, email, hashedPassword)
		if err != nil {
			log.Printf("Error inserting user: %v", err)
			http.Error(w, "Error inserting user", http.StatusInternalServerError)
			return
		}

		// Commit the transaction
		err = tx.Commit()
		if err != nil {
			log.Printf("Error committing transaction: %v", err)
			http.Error(w, "Error committing transaction", http.StatusInternalServerError)
			return
		}

		lastInsertID, err := result.LastInsertId()
		if err != nil {
			log.Printf("Error getting last inserted ID: %v", err)
		} else {
			log.Printf("Last inserted ID: %d", lastInsertID)
		}

		http.Redirect(w, r, "/home", http.StatusSeeOther)
	} else {
		t, err := template.ParseFiles("templates/register.html")
		if err != nil {
			http.Error(w, "Error parsing template", http.StatusInternalServerError)
			return
		}
		err = t.Execute(w, nil)
		if err != nil {
			http.Error(w, "Error executing template", http.StatusInternalServerError)
			return
		}
	}
}
