package forum

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func LikeHandler(w http.ResponseWriter, r *http.Request) {
	// Set the content type to JSON
	w.Header().Set("Content-Type", "application/json")

	// Get the post ID from the request URL path
	postIDStr := strings.TrimPrefix(r.URL.Path, "/like/")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		ErrorHandler(w, r, http.StatusBadRequest)
		return
	}

	// Get the session ID from the cookie
	sessionID, _ := getCookie(r, CookieName)
	var userID int
	err = Db.QueryRow("SELECT user_id FROM sessions WHERE id = ?", sessionID).Scan(&userID)
	if err != nil {
		ErrorHandler(w, r, http.StatusUnauthorized)
		return
	}

	// Check if the user has already liked the post
	var existingLikes int
	err = Db.QueryRow("SELECT COUNT(*) FROM PostLikes WHERE user_id = ? AND post_id = ?", userID, postID).Scan(&existingLikes)
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}

	if existingLikes > 0 {
		// User has already liked the post, remove their like
		_, err = Db.Exec("DELETE FROM PostLikes WHERE user_id = ? AND post_id = ?", userID, postID)
		if err != nil {
			http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
			return
		}
	} else {
		// User has not liked the post
		var existingDisikes int
		err = Db.QueryRow("SELECT COUNT(*) FROM PostDislikes WHERE user_id = ? AND post_id = ?", userID, postID).Scan(&existingDisikes)
		if err != nil {
			ErrorHandler(w, r, http.StatusInternalServerError)
			return
		}
		if existingDisikes > 0 {
			// User has already disliked the post, remove their dislike
			_, err = Db.Exec("DELETE FROM PostDislikes WHERE user_id = ? AND post_id = ?", userID, postID)
			if err != nil {
				http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
				return
			}
		}	
		_, err = Db.Exec("INSERT INTO PostLikes (user_id, post_id) VALUES (?, ?)", userID, postID)
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

	// Update the like count in the posts table
	_, err = Db.Exec("UPDATE posts SET like_count = (SELECT COUNT(*) FROM PostLikes WHERE post_id = ?) WHERE post_id = ?", postID, postID)
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

	var dislikeCount int
	err = Db.QueryRow("SELECT dislike_count FROM posts WHERE post_id = ?", postID).Scan(&dislikeCount)
	if err != nil {
		http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
		return
	}
	// Return the updated like count as JSON
	json.NewEncoder(w).Encode(map[string]int{"likeCount": likeCount, "dislikeCount":dislikeCount})
}

func CommentikeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	// Get the post ID from the request URL path
	commentIDStr := strings.TrimPrefix(r.URL.Path, "/clike/")
	CommentID, err := strconv.Atoi(commentIDStr)
	if err != nil {
		// http.Error(w, "Invalid post ID", http.StatusBadRequest)
		ErrorHandler(w, r, http.StatusBadRequest)
		return
	}

	// Get the session ID from the cookie
	sessionID, _ := getCookie(r, CookieName)
	var userID int
	err = Db.QueryRow("SELECT user_id FROM sessions WHERE id = ?", sessionID).Scan(&userID)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Check if the user has already liked the post
	var existingLikes int
	err = Db.QueryRow("SELECT COUNT(*) FROM CommentLikes WHERE user_id = ? AND comment_id = ?", userID, CommentID).Scan(&existingLikes)
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError)
		// http.Error(w, "Database error1", http.StatusInternalServerError)
		return
	}

	if existingLikes > 0 {
		// User has already liked the comment, remove their like
		_, err = Db.Exec("DELETE FROM CommentLikes WHERE user_id = ? AND comment_id = ?", userID, CommentID)
		if err != nil {
			// http.Error(w, "Database error2", http.StatusInternalServerError)
			ErrorHandler(w, r, http.StatusInternalServerError)
			return
		}
	} else {
		// User has not liked the comment, add their like
		_, err = Db.Exec("INSERT INTO CommentLikes (user_id, comment_id) VALUES (?, ?)", userID, CommentID)
		if err != nil {
			// http.Error(w, "Database error3", http.StatusInternalServerError)
			ErrorHandler(w, r, http.StatusInternalServerError)
			return
		}
	}

	//if the same user.id is on PostLikes then delete it.
	var existingDisikes int
	err = Db.QueryRow("SELECT COUNT(*) FROM CommentDislikes WHERE user_id = ? AND comment_id = ?", userID, CommentID).Scan(&existingDisikes)
	if err != nil {
		// http.Error(w, "Database error4", http.StatusInternalServerError)
		ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}
	if existingDisikes > 0 {
		// User has already liked the post, remove their like
		_, err = Db.Exec("DELETE FROM CommentDislikes WHERE user_id = ? AND comment_id = ?", userID, CommentID)
		if err != nil {
			// http.Error(w, "Database error5", http.StatusInternalServerError)
			ErrorHandler(w, r, http.StatusInternalServerError)
			return
		}
	}
	// Update the dislike count in the posts table
	_, err = Db.Exec("UPDATE comments SET dislike_count = (SELECT COUNT(*) FROM CommentDislikes WHERE comment_id = ?) WHERE comment_id = ?", CommentID, CommentID)
	if err != nil {
		// http.Error(w, "Database error6", http.StatusInternalServerError)
		ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}

	// Update the like count in the posts table
	_, err = Db.Exec("UPDATE comments SET like_count = (SELECT COUNT(*) FROM CommentLikes WHERE comment_id = ?) WHERE comment_id = ?", CommentID, CommentID)
	if err != nil {
		// http.Error(w, "Database error7", http.StatusInternalServerError)
		ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}

	// Get the post ID for this comment
	postID, err := getPostIDFromCommentID(CommentID)
	if err != nil {
		log.Printf("Error getting post ID: %v", err)
		ErrorHandler(w, r, http.StatusInternalServerError)
		// http.Error(w, "Database error8", http.StatusInternalServerError)
		return
	}

	// Get the updated like count
	var clikeCount int
	err = Db.QueryRow("SELECT like_count FROM Comments WHERE post_id = ?", postID).Scan(&clikeCount)
	if err != nil {
		http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
		return
	}

	var cdislikeCount int
	err = Db.QueryRow("SELECT dislike_count FROM Comments WHERE post_id = ?", postID).Scan(&cdislikeCount)
	if err != nil {
		http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]int{"likeCount": clikeCount, "dislikeCount":cdislikeCount})
}
