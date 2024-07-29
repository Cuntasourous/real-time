package forum

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func DislikeHandler(w http.ResponseWriter, r *http.Request) {
	// Set the content type to JSON
	w.Header().Set("Content-Type", "application/json")

	// Get the post ID from the request URL path
	postIDStr := strings.TrimPrefix(r.URL.Path, "/dislike/")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		http.Error(w, `{"error": "Invalid post ID"}`, http.StatusBadRequest)
		return
	}

	// Get the session ID from the cookie
	sessionID, _ := getCookie(r, CookieName)
	var userID int
	err = Db.QueryRow("SELECT user_id FROM sessions WHERE id = ?", sessionID).Scan(&userID)
	if err != nil {
		http.Error(w, `{"error": "User not logged in"}`, http.StatusUnauthorized)
		return
	}

	// Check if the user has already disliked the post
	var existingDislikes int
	err = Db.QueryRow("SELECT COUNT(*) FROM PostDislikes WHERE user_id = ? AND post_id = ?", userID, postID).Scan(&existingDislikes)
	if err != nil {
		http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
		return
	}

	if existingDislikes > 0 {
		// User has already disliked the post, remove their dislike
		_, err = Db.Exec("DELETE FROM PostDislikes WHERE user_id = ? AND post_id = ?", userID, postID)
		if err != nil {
			http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
			return
		}
	} else {
		// User has not disliked the post
		// If the same user.id is on PostLikes then delete it.
		var existingLikes int
		err = Db.QueryRow("SELECT COUNT(*) FROM PostLikes WHERE user_id = ? AND post_id = ?", userID, postID).Scan(&existingLikes)
		if err != nil {
			http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
			return
		}
		if existingLikes > 0 {
			// User has already liked the post, remove their like
			_, err = Db.Exec("DELETE FROM PostLikes WHERE user_id = ? AND post_id = ?", userID, postID)
			if err != nil {
				http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
				return
			}
		}

		// Update the like count in the posts table
		_, err = Db.Exec("UPDATE posts SET like_count = (SELECT COUNT(*) FROM PostLikes WHERE post_id = ?) WHERE post_id = ?", postID, postID)
		if err != nil {
			http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
			return
		}
		_, err = Db.Exec("INSERT INTO PostDislikes (user_id, post_id) VALUES (?, ?)", userID, postID)
		if err != nil {
			http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
			return
		}
	}

	// Update the dislike count in the posts table
	_, err = Db.Exec("UPDATE posts SET dislike_count = (SELECT COUNT(*) FROM PostDislikes WHERE post_id = ?) WHERE post_id = ?", postID, postID)
	if err != nil {
		http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
		return
	}

	// Get the updated dislike count
	var dislikeCount int
	err = Db.QueryRow("SELECT dislike_count FROM posts WHERE post_id = ?", postID).Scan(&dislikeCount)
	if err != nil {
		http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
		return
	}

	// Get the updated like count
	var likeCount int
	err = Db.QueryRow("SELECT like_count FROM posts WHERE post_id = ?", postID).Scan(&likeCount)
	if err != nil {
		http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
		return
	}

	// Return the updated dislike count as JSON
	json.NewEncoder(w).Encode(map[string]int{"dislikeCount": dislikeCount, "likeCount": likeCount})
}

func CommentDislikeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	commentIDStr := strings.TrimPrefix(r.URL.Path, "/cdislike/")
	commentID, err := strconv.Atoi(commentIDStr)
	if err != nil {
		ErrorHandler(w, r, http.StatusBadRequest)
		return
	}

	sessionID, _ := getCookie(r, CookieName)
	var userID int
	err = Db.QueryRow("SELECT user_id FROM sessions WHERE id = ?", sessionID).Scan(&userID)
	if err != nil {
		log.Printf("Session retrieval error: %v", err) // Add this line
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	var existingDislikes int
	err = Db.QueryRow("SELECT COUNT(*) FROM CommentDislikes WHERE user_id = ? AND comment_id = ?", userID, commentID).Scan(&existingDislikes)
	if err != nil {
		log.Printf("Dislike count retrieval error: %v", err) // Add this line
		ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}

	if existingDislikes > 0 {
		_, err = Db.Exec("DELETE FROM CommentDislikes WHERE user_id = ? AND comment_id = ?", userID, commentID)
		if err != nil {
			log.Printf("Error removing dislike: %v", err) // Add this line
			ErrorHandler(w, r, http.StatusInternalServerError)
			return
		}
	} else {
		_, err = Db.Exec("INSERT INTO CommentDislikes (user_id, comment_id) VALUES (?, ?)", userID, commentID)
		if err != nil {
			log.Printf("Error adding dislike: %v", err) // Add this line
			http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
			return
		}
	}

	var existingLikes int
	err = Db.QueryRow("SELECT COUNT(*) FROM CommentLikes WHERE user_id = ? AND comment_id = ?", userID, commentID).Scan(&existingLikes)
	if err != nil {
		log.Printf("Like count retrieval error: %v", err) // Add this line
		http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
		return
	}
	if existingLikes > 0 {
		_, err = Db.Exec("DELETE FROM CommentLikes WHERE user_id = ? AND comment_id = ?", userID, commentID)
		if err != nil {
			log.Printf("Error removing like: %v", err) // Add this line
			http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
			return
		}
	}

	_, err = Db.Exec("UPDATE comments SET like_count = (SELECT COUNT(*) FROM CommentLikes WHERE comment_id = ?) WHERE comment_id = ?", commentID, commentID)
	if err != nil {
		log.Printf("Error updating like count: %v", err) // Add this line
		ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}

	_, err = Db.Exec("UPDATE comments SET dislike_count = (SELECT COUNT(*) FROM CommentDislikes WHERE comment_id = ?) WHERE comment_id = ?", commentID, commentID)
	if err != nil {
		log.Printf("Error updating dislike count: %v", err) // Add this line
		ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}

	var clikeCount int
	err = Db.QueryRow("SELECT like_count FROM Comments WHERE comment_id = ?", commentID).Scan(&clikeCount)
	if err != nil {
		log.Printf("Error retrieving like count: %v", err) // Add this line
		http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
		return
	}

	var cdislikeCount int
	err = Db.QueryRow("SELECT dislike_count FROM Comments WHERE comment_id = ?", commentID).Scan(&cdislikeCount)
	if err != nil {
		log.Printf("Error retrieving dislike count: %v", err) // Add this line
		http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]int{"likeCount": clikeCount, "dislikeCount": cdislikeCount})
}

func getPostIDFromCommentID(commentID int) (int, error) {
	var postID int
	err := Db.QueryRow("SELECT post_id FROM Comments WHERE comment_id = ?", commentID).Scan(&postID)
	if err != nil {
		return 0, err
	}
	return postID, nil
}
