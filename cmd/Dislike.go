package forum

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func DislikeHandler(w http.ResponseWriter, r *http.Request) {
	// Get the post ID from the request URL path
	postIDStr := strings.TrimPrefix(r.URL.Path, "/dislike/")
	postID, err := strconv.Atoi(postIDStr)
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

	// Check if the user has already disliked the post
	var existingDislikes int
	err = Db.QueryRow("SELECT COUNT(*) FROM PostDislikes WHERE user_id = ? AND post_id = ?", userID, postID).Scan(&existingDislikes)
	if err != nil {
		// http.Error(w, "Database error", http.StatusInternalServerError)
		ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}

	//method 2: by counting the UserIDs per postID
	if existingDislikes > 0 {
		// User has already disliked the post, remove their like
		_, err = Db.Exec("DELETE FROM PostDislikes WHERE user_id = ? AND post_id = ?", userID, postID)
		if err != nil {
			// http.Error(w, "Database error", http.StatusInternalServerError)
			ErrorHandler(w, r, http.StatusInternalServerError)
			return
		}
	} else {
		// User has not disliked the post, add their dislike
		_, err = Db.Exec("INSERT INTO PostDislikes (user_id, post_id) VALUES (?, ?)", userID, postID)
		if err != nil {
			// http.Error(w, "Database error", http.StatusInternalServerError)
			ErrorHandler(w, r, http.StatusInternalServerError)
			return
		}
	}

	//if the same user.id is on PostLikes then delete it.
	var existingLikes int
	err = Db.QueryRow("SELECT COUNT(*) FROM PostLikes WHERE user_id = ? AND post_id = ?", userID, postID).Scan(&existingLikes)
	if err != nil {
		// http.Error(w, "Database error", http.StatusInternalServerError)
		ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}
	if existingLikes > 0 {
		// User has already liked the post, remove their like
		_, err = Db.Exec("DELETE FROM PostLikes WHERE user_id = ? AND post_id = ?", userID, postID)
		if err != nil {
			// http.Error(w, "Database error", http.StatusInternalServerError)
			ErrorHandler(w, r, http.StatusInternalServerError)
			return
		}
	}

	// Update the like count in the posts table
	_, err = Db.Exec("UPDATE posts SET like_count = (SELECT COUNT(*) FROM PostLikes WHERE post_id = ?) WHERE post_id = ?", postID, postID)
	if err != nil {
		// http.Error(w, "Database error", http.StatusInternalServerError)
		ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}

	// Update the dislike count in the posts table
	_, err = Db.Exec("UPDATE posts SET dislike_count = (SELECT COUNT(*) FROM PostDislikes WHERE post_id = ?) WHERE post_id = ?", postID, postID)
	if err != nil {
		// http.Error(w, "Database error", http.StatusInternalServerError)
		ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}

	// Redirect to the homepage
	http.Redirect(w, r, "/", http.StatusFound)
}

func CommentDislikeHandler(w http.ResponseWriter, r *http.Request) {
	// Get the post ID from the request URL path
	commentIDStr := strings.TrimPrefix(r.URL.Path, "/cdislike/")
	commentID, err := strconv.Atoi(commentIDStr)
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

	// Check if the user has already disliked the post
	var existingDislikes int
	err = Db.QueryRow("SELECT COUNT(*) FROM CommentDislikes WHERE user_id = ? AND comment_id = ?", userID, commentID).Scan(&existingDislikes)
	if err != nil {
		// http.Error(w, "Database error1", http.StatusInternalServerError)
		ErrorHandler(w, r, http.StatusInternalServerError)
		// fmt.Println("err:", err)
		return
	}

	//method 2: by counting the UserIDs per commentID
	if existingDislikes > 0 {
		// User has already disliked the post, remove their like
		_, err = Db.Exec("DELETE FROM CommentDislikes WHERE user_id = ? AND comment_id = ?", userID, commentID)
		if err != nil {
			// http.Error(w, "Database error2", http.StatusInternalServerError)
			ErrorHandler(w, r, http.StatusInternalServerError)
			return
		}
	} else {
		// User has not disliked the post, add their dislike
		_, err = Db.Exec("INSERT INTO CommentDislikes (user_id, comment_id) VALUES (?, ?)", userID, commentID)
		if err != nil {
			// http.Error(w, "Database error3", http.StatusInternalServerError)
			ErrorHandler(w, r, http.StatusInternalServerError)
			return
		}
	}

	//if the same user.id is on PostLikes then delete it.
	var existingLikes int
	err = Db.QueryRow("SELECT COUNT(*) FROM CommentLikes WHERE user_id = ? AND comment_id = ?", userID, commentID).Scan(&existingLikes)
	if err != nil {
		// http.Error(w, "Database error4", http.StatusInternalServerError)
		ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}
	if existingLikes > 0 {
		// User has already liked the post, remove their like
		_, err = Db.Exec("DELETE FROM CommentLikes WHERE user_id = ? AND comment_id = ?", userID, commentID)
		if err != nil {
			// http.Error(w, "Database error5", http.StatusInternalServerError)
			ErrorHandler(w, r, http.StatusInternalServerError)
			return
		}
	}

	// Update the like count in the posts table
	_, err = Db.Exec("UPDATE comments SET like_count = (SELECT COUNT(*) FROM CommentLikes WHERE comment_id = ?) WHERE comment_id = ?", commentID, commentID)
	if err != nil {
		// http.Error(w, "Database error", http.StatusInternalServerError)
		ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}

	// Update the dislike count in the posts table
	_, err = Db.Exec("UPDATE comments SET dislike_count = (SELECT COUNT(*) FROM CommentDislikes WHERE comment_id = ?) WHERE comment_id = ?", commentID, commentID)
	if err != nil {
		// http.Error(w, "Database error", http.StatusInternalServerError)
		ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}

	// Get the post ID for this comment
	postID, err := getPostIDFromCommentID(commentID)
	if err != nil {
		log.Printf("Error getting post ID: %v", err)
		// http.Error(w, "Database error", http.StatusInternalServerError)
		ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}

	// Redirect to the post page
	http.Redirect(w, r, fmt.Sprintf("/view_post/%d", postID), http.StatusFound)
}

func getPostIDFromCommentID(commentID int) (int, error) {
	var postID int
	err := Db.QueryRow("SELECT post_id FROM Comments WHERE comment_id = ?", commentID).Scan(&postID)
	if err != nil {
		return 0, err
	}
	return postID, nil
}

func DislikeHandler2(w http.ResponseWriter, r *http.Request) {
	// Get the post ID from the request URL path
	postIDStr := strings.TrimPrefix(r.URL.Path, "/dislike2/")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
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

	// Check if the user has already disliked the post
	var existingDislikes int
	err = Db.QueryRow("SELECT COUNT(*) FROM PostDislikes WHERE user_id = ? AND post_id = ?", userID, postID).Scan(&existingDislikes)
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	//method 2: by counting the UserIDs per postID
	if existingDislikes > 0 {
		// User has already disliked the post, remove their like
		_, err = Db.Exec("DELETE FROM PostDislikes WHERE user_id = ? AND post_id = ?", userID, postID)
		if err != nil {
			ErrorHandler(w, r, http.StatusInternalServerError)
			// http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
	} else {
		// User has not disliked the post, add their dislike
		_, err = Db.Exec("INSERT INTO PostDislikes (user_id, post_id) VALUES (?, ?)", userID, postID)
		if err != nil {
			// http.Error(w, "Database error", http.StatusInternalServerError)
			ErrorHandler(w, r, http.StatusInternalServerError)
			return
		}
	}

	//if the same user.id is on PostLikes then delete it.
	var existingLikes int
	err = Db.QueryRow("SELECT COUNT(*) FROM PostLikes WHERE user_id = ? AND post_id = ?", userID, postID).Scan(&existingLikes)
	if err != nil {
		// http.Error(w, "Database error", http.StatusInternalServerError)
		ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}
	if existingLikes > 0 {
		// User has already liked the post, remove their like
		_, err = Db.Exec("DELETE FROM PostLikes WHERE user_id = ? AND post_id = ?", userID, postID)
		if err != nil {
			// http.Error(w, "Database error", http.StatusInternalServerError)
			ErrorHandler(w, r, http.StatusInternalServerError)
			return
		}
	}

	// Update the like count in the posts table
	_, err = Db.Exec("UPDATE posts SET like_count = (SELECT COUNT(*) FROM PostLikes WHERE post_id = ?) WHERE post_id = ?", postID, postID)
	if err != nil {
		// http.Error(w, "Database error", http.StatusInternalServerError)
		ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}

	// Update the dislike count in the posts table
	_, err = Db.Exec("UPDATE posts SET dislike_count = (SELECT COUNT(*) FROM PostDislikes WHERE post_id = ?) WHERE post_id = ?", postID, postID)
	if err != nil {
		// http.Error(w, "Database error", http.StatusInternalServerError)
		ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}

	// After updating the dislike count, redirect back to the view_post page
	http.Redirect(w, r, fmt.Sprintf("/view_post/%d", postID), http.StatusFound)
}
