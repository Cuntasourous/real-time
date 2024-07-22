package forum

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func getCommentsByPostID(postID int) ([]Comment, error) {
	var comments []Comment
	rows, err := Db.Query("SELECT c.comment_id, c.user_id, c.comment_text, c.like_count, c.dislike_count, u.username FROM Comments c JOIN Users u ON c.user_id = u.user_id WHERE c.post_id = ?", postID)
	if err != nil {
		log.Printf("Error getting comments: %v", err)
		return nil, err // Return the error
	}
	defer rows.Close()

	for rows.Next() {
		var comment Comment
		err := rows.Scan(&comment.CommentID, &comment.UserID, &comment.CommentText, &comment.LikeCount, &comment.DislikeCount, &comment.Username)
		if err != nil {
			log.Printf("Error scanning comment: %v", err)
			return nil, err // Return the error
		}
		comments = append(comments, comment)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating over comments: %v", err)
		return nil, err // Return the error
	}

	return comments, nil
}

func handleAddComment(w http.ResponseWriter, r *http.Request, postID int) {
	// Extract post_id from the URL
	// postID, err := getPostIDFromURL(r)
	// if err != nil {
	// 	http.Error(w, "Invalid post ID", http.StatusBadRequest)
	// 	return
	// }

	// Get comment text from the form
	commentText := r.FormValue("comment_text")
	if commentText == "" {
		log.Println("there is no comment text")
	}

	// Start a transaction
	tx, err := Db.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}
	defer tx.Rollback() // Roll back the transaction if it's not committed

	// Insert the category into the Categories table
	result, err := tx.Exec("INSERT INTO Comments(comment_text) VALUES(?)", commentText)
	if err != nil {
		log.Printf("Error inserting comment: %v", err)
		// http.Error(w, "Error creating comment", http.StatusInternalServerError)
		ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}

	// Get the last inserted ID
	lastInsertID, err := result.LastInsertId()
	if err != nil {
		log.Printf("Error getting last inserted ID: %v", err)
	} else {
		log.Printf("comment created with ID: %d", lastInsertID)
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		// http.Error(w, "Error creating comment", http.StatusInternalServerError)
		ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}

	log.Println("comment created successfully")

	// Redirect back to the post page
	http.Redirect(w, r, fmt.Sprintf("/view_post/%d", postID), http.StatusSeeOther)
}

func HandleAddCommentAJAX(w http.ResponseWriter, r *http.Request) {
	if !isAuthenticated(r) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get the session ID from the cookie
	sessionID, _ := getCookie(r, CookieName)
	var userID int
	err := Db.QueryRow("SELECT user_id FROM sessions WHERE id = ?", sessionID).Scan(&userID)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}

	var username string
	err = Db.QueryRow("SELECT username FROM users WHERE user_id = ?", userID).Scan(&username)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}

	if r.Method != http.MethodPost {
		// http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		ErrorHandler(w, r, http.StatusBadRequest)
		return
	}

	postID, err := getPostIDFromURL(r)
	if err != nil {
		ErrorHandler(w, r, http.StatusBadRequest)
		return
	}

	commentText := r.FormValue("comment_text")
	if commentText == "" {
		// http.Error(w, "Comment text is required", http.StatusBadRequest)
		ErrorHandler(w, r, http.StatusBadRequest)
		return
	}

	// Insert the comment into the database using SQLite's datetime('now') function
	result, err := Db.Exec("INSERT INTO Comments(post_id, user_id, comment_text, comment_date) VALUES(?, ?, ?, datetime('now'))", postID, userID, commentText)
	if err != nil {
		log.Printf("Error inserting comment: %v", err)
		ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}

	commentID, _ := result.LastInsertId()

	// Fetch the newly created comment
	var comment Comment
	err = Db.QueryRow(`
        SELECT c.comment_id, c.post_id, c.user_id, c.comment_text, c.comment_date, c.like_count, c.dislike_count, u.username 
        FROM Comments c 
        JOIN Users u ON c.user_id = u.user_id 
        WHERE c.comment_id = ?`, commentID).Scan(
		&comment.CommentID, &comment.PostID, &comment.UserID, &comment.CommentText, &comment.CommentDate, &comment.LikeCount, &comment.DislikeCount, &comment.Username)
	if err != nil {
		log.Printf("Error fetching new comment: %v", err)
		ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}

	// Return the comment data as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comment)
}
