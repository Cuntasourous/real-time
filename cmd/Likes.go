package forum

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func LikeHandler(w http.ResponseWriter, r *http.Request) {
	// Get the post ID from the request URL path
	postIDStr := strings.TrimPrefix(r.URL.Path, "/like/")
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

	// Check if the user has already liked the post
	var existingLikes int
	err = Db.QueryRow("SELECT COUNT(*) FROM PostLikes WHERE user_id = ? AND post_id = ?", userID, postID).Scan(&existingLikes)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	//method 1: increment 1 and subtract 1
	// if existingLikes > 0 {
	// 	// User has already liked the post, remove their like
	// 	_, err = Db.Exec("DELETE FROM PostLikes WHERE user_id = ? AND post_id = ?", userID, postID)
	// 	if err != nil {
	// 		http.Error(w, "Database error", http.StatusInternalServerError)
	// 		return
	// 	}

	// 	// Update the like count in the posts table
	// 	_, err = Db.Exec("UPDATE posts SET like_count = like_count - 1 WHERE post_id = ?", postID)
	// 	if err != nil {
	// 		http.Error(w, "Database error", http.StatusInternalServerError)
	// 		return
	// 	}
	// } else {
	// 	// User has not liked the post, add their like
	// 	_, err = Db.Exec("INSERT INTO PostLikes (user_id, post_id) VALUES (?, ?)", userID, postID)
	// 	if err != nil {
	// 		http.Error(w, "Database error", http.StatusInternalServerError)
	// 		return
	// 	}

	// 	// Update the like count in the posts table
	// 	_, err = Db.Exec("UPDATE posts SET like_count = like_count + 1 WHERE post_id = ?", postID)
	// 	if err != nil {
	// 		http.Error(w, "Database error", http.StatusInternalServerError)
	// 		return
	// 	}
	// }

	//method 2: by counting the UserIDs per postID

	if existingLikes > 0 {
		// User has already liked the post, remove their like
		_, err = Db.Exec("DELETE FROM PostLikes WHERE user_id = ? AND post_id = ?", userID, postID)
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
	} else {
		// User has not liked the post, add their like
		_, err = Db.Exec("INSERT INTO PostLikes (user_id, post_id) VALUES (?, ?)", userID, postID)
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
	}

	//if the same user.id is on PostLikes then delete it.
	var existingDisikes int
	err = Db.QueryRow("SELECT COUNT(*) FROM PostDislikes WHERE user_id = ? AND post_id = ?", userID, postID).Scan(&existingDisikes)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if existingDisikes > 0 {
		// User has already liked the post, remove their like
		_, err = Db.Exec("DELETE FROM PostDislikes WHERE user_id = ? AND post_id = ?", userID, postID)
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
	}
	// Update the dislike count in the posts table
	_, err = Db.Exec("UPDATE posts SET dislike_count = (SELECT COUNT(*) FROM PostDislikes WHERE post_id = ?) WHERE post_id = ?", postID, postID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Update the like count in the posts table
	_, err = Db.Exec("UPDATE posts SET like_count = (SELECT COUNT(*) FROM PostLikes WHERE post_id = ?) WHERE post_id = ?", postID, postID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Redirect to the homepage
	http.Redirect(w, r, "/", http.StatusFound)
}

func CommentikeHandler(w http.ResponseWriter, r *http.Request) {
	// Get the post ID from the request URL path
	commentIDStr := strings.TrimPrefix(r.URL.Path, "/clike/")
	CommentID, err := strconv.Atoi(commentIDStr)
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

	// Check if the user has already liked the post
	var existingLikes int
	err = Db.QueryRow("SELECT COUNT(*) FROM CommentLikes WHERE user_id = ? AND comment_id = ?", userID, CommentID).Scan(&existingLikes)
	if err != nil {
		http.Error(w, "Database error1", http.StatusInternalServerError)
		return
	}

	//method 2: by counting the UserIDs per postID

	if existingLikes > 0 {
		// User has already liked the comment, remove their like
		_, err = Db.Exec("DELETE FROM CommentLikes WHERE user_id = ? AND comment_id = ?", userID, CommentID)
		if err != nil {
			http.Error(w, "Database error2", http.StatusInternalServerError)
			return
		}
	} else {
		// User has not liked the comment, add their like
		_, err = Db.Exec("INSERT INTO CommentLikes (user_id, comment_id) VALUES (?, ?)", userID, CommentID)
		if err != nil {
			http.Error(w, "Database error3", http.StatusInternalServerError)
			return
		}
	}

	//if the same user.id is on PostLikes then delete it.
	var existingDisikes int
	err = Db.QueryRow("SELECT COUNT(*) FROM CommentDislikes WHERE user_id = ? AND comment_id = ?", userID, CommentID).Scan(&existingDisikes)
	if err != nil {
		http.Error(w, "Database error4", http.StatusInternalServerError)
		return
	}
	if existingDisikes > 0 {
		// User has already liked the post, remove their like
		_, err = Db.Exec("DELETE FROM CommentDislikes WHERE user_id = ? AND comment_id = ?", userID, CommentID)
		if err != nil {
			http.Error(w, "Database error5", http.StatusInternalServerError)
			return
		}
	}
	// Update the dislike count in the posts table
	_, err = Db.Exec("UPDATE comments SET dislike_count = (SELECT COUNT(*) FROM CommentDislikes WHERE comment_id = ?) WHERE comment_id = ?", CommentID, CommentID)
	if err != nil {
		http.Error(w, "Database error6", http.StatusInternalServerError)
		return
	}

	// Update the like count in the posts table
	_, err = Db.Exec("UPDATE comments SET like_count = (SELECT COUNT(*) FROM CommentLikes WHERE comment_id = ?) WHERE comment_id = ?", CommentID, CommentID)
	if err != nil {
		http.Error(w, "Database error7", http.StatusInternalServerError)
		return
	}

	// Get the post ID for this comment
	postID, err := getPostIDFromCommentID(CommentID)
	if err != nil {
		log.Printf("Error getting post ID: %v", err)
		http.Error(w, "Database error8", http.StatusInternalServerError)
		return
	}

	// Redirect to the post page
	http.Redirect(w, r, fmt.Sprintf("/view_post/%d", postID), http.StatusFound)
}

func LikeHandler2(w http.ResponseWriter, r *http.Request) {
	// Get the post ID from the request URL path
	postIDStr := strings.TrimPrefix(r.URL.Path, "/like2/")
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

	// Check if the user has already liked the post
	var existingLikes int
	err = Db.QueryRow("SELECT COUNT(*) FROM PostLikes WHERE user_id = ? AND post_id = ?", userID, postID).Scan(&existingLikes)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	//method 2: by counting the UserIDs per postID

	if existingLikes > 0 {
		// User has already liked the post, remove their like
		_, err = Db.Exec("DELETE FROM PostLikes WHERE user_id = ? AND post_id = ?", userID, postID)
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
	} else {
		// User has not liked the post, add their like
		_, err = Db.Exec("INSERT INTO PostLikes (user_id, post_id) VALUES (?, ?)", userID, postID)
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
	}

	//if the same user.id is on PostLikes then delete it.
	var existingDisikes int
	err = Db.QueryRow("SELECT COUNT(*) FROM PostDislikes WHERE user_id = ? AND post_id = ?", userID, postID).Scan(&existingDisikes)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if existingDisikes > 0 {
		// User has already liked the post, remove their like
		_, err = Db.Exec("DELETE FROM PostDislikes WHERE user_id = ? AND post_id = ?", userID, postID)
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
	}
	// Update the dislike count in the posts table
	_, err = Db.Exec("UPDATE posts SET dislike_count = (SELECT COUNT(*) FROM PostDislikes WHERE post_id = ?) WHERE post_id = ?", postID, postID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Update the like count in the posts table
	_, err = Db.Exec("UPDATE posts SET like_count = (SELECT COUNT(*) FROM PostLikes WHERE post_id = ?) WHERE post_id = ?", postID, postID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// After updating the like count, redirect back to the view_post page
	http.Redirect(w, r, fmt.Sprintf("/view_post/%d", postID), http.StatusFound)
}
