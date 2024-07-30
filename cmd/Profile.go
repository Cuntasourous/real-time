package forum

import (
	"html/template"
	"net/http"
	"fmt"
)

func ViewProfileHandler(w http.ResponseWriter, r *http.Request) {
	if !isAuthenticated(r, w) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	sessionID, _ := getCookie(r, w, CookieName)
	var userID int
	err := Db.QueryRow("SELECT user_id FROM sessions WHERE id = ?", sessionID).Scan(&userID)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	profile, err := getUserProfile(userID)
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}

	t, err := template.ParseFiles("templates/view_profile.html")
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}

	err = t.Execute(w, profile)
	if err != nil {
		fmt.Println("here", err)
		ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}
}

func getUserProfile(userID int) (UserProfile, error) {
	var profile UserProfile

	// Fetch user details
	var user User
	err := Db.QueryRow("SELECT username, email, date_created FROM users WHERE user_id = ?", userID).
		Scan(&user.Username, &user.Email, &user.DateCreated)
	if err != nil {
		return profile, err
	}
	profile.Username = user.Username
	profile.Email = user.Email
	profile.DateCreated = user.DateCreated

	// Fetch user's posts and count
	rows, err := Db.Query("SELECT post_id, user_id, post_text, post_date, like_count, dislike_count FROM Posts WHERE user_id = ?", userID)
	if err != nil {
		return profile, err
	}
	defer rows.Close()

	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.PostID, &post.UserID, &post.PostText, &post.PostDate, &post.LikeCount, &post.DislikeCount); err != nil {
			return profile, err
		}
		profile.Posts = append(profile.Posts, post)
	}
	profile.PostCount = len(profile.Posts)

	// Fetch user's comments and count
	rows, err = Db.Query("SELECT comment_id, post_id, user_id, comment_text, comment_date, like_count, dislike_count FROM Comments WHERE user_id = ?", userID)
	if err != nil {
		return profile, err
	}
	defer rows.Close()

	for rows.Next() {
		var comment Comment
		if err := rows.Scan(&comment.CommentID, &comment.PostID, &comment.UserID, &comment.CommentText, &comment.CommentDate, &comment.LikeCount, &comment.DislikeCount); err != nil {
			return profile, err
		}
		profile.Comments = append(profile.Comments, comment)
	}
	profile.CommentCount = len(profile.Comments)

	// Fetch user's liked posts and count
	rows, err = Db.Query(`
        SELECT p.post_id, p.user_id, p.post_text, p.post_date, p.like_count, p.dislike_count
        FROM Posts p
        JOIN PostLikes pl ON p.post_id = pl.post_id
        WHERE pl.user_id = ?`, userID)
	if err != nil {
		return profile, err
	}
	defer rows.Close()

	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.PostID, &post.UserID, &post.PostText, &post.PostDate, &post.LikeCount, &post.DislikeCount); err != nil {
			return profile, err
		}
		profile.LikedPosts = append(profile.LikedPosts, post)
	}
	profile.LikedPostCount = len(profile.LikedPosts)

	// Fetch user's disliked posts and count
	rows, err = Db.Query(`
        SELECT p.post_id, p.user_id, p.post_text, p.post_date, p.like_count, p.dislike_count
        FROM Posts p
        JOIN PostDislikes pdl ON p.post_id = pdl.post_id
        WHERE pdl.user_id = ?`, userID)
	if err != nil {
		return profile, err
	}
	defer rows.Close()

	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.PostID, &post.UserID, &post.PostText, &post.PostDate, &post.LikeCount, &post.DislikeCount); err != nil {
			return profile, err
		}
		profile.DislikedPosts = append(profile.DislikedPosts, post)
	}
	profile.DislikeCount = len(profile.DislikedPosts)

	return profile, nil
}
