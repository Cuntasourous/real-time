package forum

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func PostHandler(w http.ResponseWriter, r *http.Request) {
	if !isAuthenticated(r, w) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get the session ID from the cookie
	sessionID, _ := getCookie(r, w,CookieName)
	var userID int
	err := Db.QueryRow("SELECT user_id FROM sessions WHERE id = ?", sessionID).Scan(&userID)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	var username string
	err = Db.QueryRow("SELECT username FROM users WHERE user_id = ?", userID).Scan(&username)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	categories, err := getCategories()
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}

	if r.Method == "POST" {
		post_text := r.FormValue("post_text")
		err := r.ParseForm()
		if err != nil {
			ErrorHandler(w, r, http.StatusBadRequest)
			return
		}

		selectedCategories := r.Form["categories"]

		if len(post_text) > 128 {
			data := struct {
				LoggedInUser string
				Categories   []Category
				ErrorMessage string
			}{
				LoggedInUser: username,
				Categories:   categories,
				ErrorMessage: "Maximum number of charachters is 128 words",
			}
			t, err := template.ParseFiles("templates/create_post.html")
			if err != nil {
				ErrorHandler(w, r, http.StatusInternalServerError)
				return
			}
			err = t.Execute(w, data)
			if err != nil {
				ErrorHandler(w, r, http.StatusInternalServerError)
			}
			return
		}

		if post_text == "" || len(selectedCategories) == 0 {
			data := struct {
				LoggedInUser string
				Categories   []Category
				ErrorMessage string
			}{
				LoggedInUser: username,
				Categories:   categories,
				ErrorMessage: "Please add some text and select at least one category.",
			}
			t, err := template.ParseFiles("templates/create_post.html")
			if err != nil {
				ErrorHandler(w, r, http.StatusInternalServerError)
				return
			}
			err = t.Execute(w, data)
			if err != nil {
				ErrorHandler(w, r, http.StatusInternalServerError)
			}
			return
		}

		err = createPost(userID, post_text, selectedCategories)
		if err != nil {
			ErrorHandler(w, r, http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return
	}

	// GET request handling
	data := struct {
		LoggedInUser string
		Categories   []Category
		ErrorMessage string
	}{
		LoggedInUser: username,
		Categories:   categories,
		ErrorMessage: "",
	}

	t, err := template.ParseFiles("templates/create_post.html")
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}
	err = t.Execute(w, data)
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError)
	}
}

func getPostByID(postID int) (Post, error) {
	var post Post
	err := Db.QueryRow(`
        SELECT p.post_id, p.user_id, u.username, p.post_text, p.post_date, p.like_count, p.dislike_count 
        FROM Posts p
        JOIN Users u ON p.user_id = u.user_id
        WHERE p.post_id = ?`, postID).Scan(
		&post.PostID, &post.UserID, &post.Username, &post.PostText, &post.PostDate, &post.LikeCount, &post.DislikeCount)
	if err != nil {
		return post, err
	}
	return post, nil
}

// Define the struct with correct embedding
type PostLikedOrNot struct {
	Post
	IsLiked    bool
	IsDisliked bool
}

func HandleViewPost(w http.ResponseWriter, r *http.Request) {
	// Extract the post_id from the URL
	postID, err := getPostIDFromURL(r)
	if err != nil {
			ErrorHandler(w, r, http.StatusBadRequest)
			return
	}

	if r.Method == http.MethodPost {
			// Handle the addition of a new comment
			var requestData map[string]string
			if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
					ErrorHandler(w, r, http.StatusBadRequest)
					return
			}

			commentText, ok := requestData["comment_text"]
			if !ok || commentText == "" {
					ErrorHandler(w, r, http.StatusBadRequest)
					return
			}

			// Add the comment to the database (pseudo code)
			newComment, err := addComment(w, r, postID, commentText)
			if err != nil {
					ErrorHandler(w, r, http.StatusInternalServerError)
					return
			}

			// Return the new comment as JSON
			response := map[string]interface{}{
					"comment": newComment,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
	}

	// Handle GET requests
	sessionID, _ := getCookie(r, w, CookieName)
	var userID int
	err = Db.QueryRow("SELECT user_id FROM sessions WHERE id = ?", sessionID).Scan(&userID)
	if err != nil {
			fmt.Println("guest") // Log for debugging
	}

	var username string
	err = Db.QueryRow("SELECT username FROM users WHERE user_id = ?", userID).Scan(&username)
	if err != nil {
			username = "" // No username if not found
	}

	// Fetch the post data
	post, err := getPostByID(postID)
	if err != nil {
			ErrorHandler(w, r, http.StatusNotFound)
			return
	}

	// Fetch categories for the post
	categories, err := getCategoriesByPostID(postID)
	if err != nil {
			ErrorHandler(w, r, http.StatusInternalServerError)
			return
	}

	// Fetch comments for the post
	comments, err := getCommentsByPostID(postID)
	if err != nil {
			ErrorHandler(w, r, http.StatusInternalServerError)
			return
	}

	for i := 0; i < len(comments); i++ {
			var isLiked, isDisliked bool

			err = Db.QueryRow("SELECT EXISTS(SELECT 1 FROM CommentLikes WHERE comment_id = ? AND user_id = ?)", comments[i].CommentID, userID).Scan(&isLiked)
			if err != nil {
					ErrorHandler(w, r, http.StatusInternalServerError)
					return
			}

			err = Db.QueryRow("SELECT EXISTS(SELECT 1 FROM CommentDislikes WHERE comment_id = ? AND user_id = ?)", comments[i].CommentID, userID).Scan(&isDisliked)
			if err != nil {
					ErrorHandler(w, r, http.StatusInternalServerError)
					return
			}

			comments[i].IsLiked = isLiked
			comments[i].IsDisliked = isDisliked
	}

	// Fetch popular categories
	popularCategories, err := getPopularCategories()
	if err != nil {
			log.Printf("Error fetching popular categories: %v", err)
			popularCategories = []PopularCategory{} // Provide an empty slice on error
	}

	// Check if the post is liked or disliked by the user
	var isLiked, isDisliked bool
	err = Db.QueryRow("SELECT EXISTS(SELECT 1 FROM PostLikes WHERE post_id = ? AND user_id = ?)", post.PostID, userID).Scan(&isLiked)
	if err != nil {
			ErrorHandler(w, r, http.StatusInternalServerError)
			return
	}
	err = Db.QueryRow("SELECT EXISTS(SELECT 1 FROM PostDislikes WHERE post_id = ? AND user_id = ?)", post.PostID, userID).Scan(&isDisliked)
	if err != nil {
			ErrorHandler(w, r, http.StatusInternalServerError)
			return
	}

	postLikedOrNot := PostLikedOrNot{
			Post:       post,
			IsLiked:    isLiked,
			IsDisliked: isDisliked,
	}

	data := map[string]interface{}{
			"Post":            post,
			"Categories":      categories,
			"Comments":        comments,
			"LoggedInUser":    username,
			"PopularCategory": popularCategories,
			"PostLikedOrNot":  postLikedOrNot,
	}

	// Parse the template file
	t, err := template.ParseFiles("templates/view_post.html")
	if err != nil {
			fmt.Println("here", err)
			ErrorHandler(w, r, http.StatusInternalServerError)
			return
	}

	// Execute the template with the data
	err = t.Execute(w, data)
	if err != nil {
			ErrorHandler(w, r, http.StatusInternalServerError)
			return
	}
}


func addComment(w http.ResponseWriter, r *http.Request, postID int, commentText string) (Comment, error) {
	var newComment Comment

	sessionID, _ := getCookie(r, w, CookieName)
	var userID int
	err := Db.QueryRow("SELECT user_id FROM sessions WHERE id = ?", sessionID).Scan(&userID)
	if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return newComment, err
	}

	result, err := Db.Exec("INSERT INTO Comments (user_id, post_id, comment_text) VALUES (?, ?, ?)", userID, postID, commentText)
	if err != nil {
			return newComment, err
	}

	// Get the last inserted ID
	commentID, err := result.LastInsertId()
	if err != nil {
			return newComment, err
	}

	var existingLikes int
	err = Db.QueryRow("SELECT COUNT(*) FROM CommentLikes WHERE user_id = ? AND comment_id = ?", userID, commentID).Scan(&existingLikes)
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError)
		// http.Error(w, "Database error1", http.StatusInternalServerError)
	}

	var existingDisikes int
	err = Db.QueryRow("SELECT COUNT(*) FROM CommentDislikes WHERE user_id = ? AND comment_id = ?", userID, commentID).Scan(&existingDisikes)
	if err != nil {
		// http.Error(w, "Database error4", http.StatusInternalServerError)
		ErrorHandler(w, r, http.StatusInternalServerError)
	}

	// Fetch the newly added comment from the database
	err = Db.QueryRow(`
			SELECT c.comment_id, c.comment_text, u.username, 
						 COALESCE(l.like_count, 0) as like_count, 
						 COALESCE(d.dislike_count, 0) as dislike_count,
						 c.user_id
			FROM Comments c
			JOIN users u ON c.user_id = u.user_id
			LEFT JOIN (SELECT comment_id, COUNT(*) as like_count FROM CommentLikes GROUP BY comment_id) l ON c.comment_id = l.comment_id
			LEFT JOIN (SELECT comment_id, COUNT(*) as dislike_count FROM CommentDislikes GROUP BY comment_id) d ON c.comment_id = d.comment_id
			WHERE c.comment_id = ?`, commentID).Scan(
			&newComment.CommentID,
			&newComment.CommentText,
			&newComment.Username,
			&existingLikes,
			&existingDisikes,
			&newComment.UserID,
	)
	if err != nil {
			return newComment, err
	}

	newComment.DislikeCount = existingDisikes
	newComment.LikeCount = existingLikes

	// Return the newly added comment
	return newComment, nil
}


func getPostIDFromURL(r *http.Request) (int, error) {
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		return 0, fmt.Errorf("invalid URL path")
	}
	postID, err := strconv.Atoi(pathParts[len(pathParts)-1])
	if err != nil {
		return 0, fmt.Errorf("invalid post ID")
	}
	return postID, nil
}

func createPost(userID int, postText string, selectedCategories []string) error {
	tx, err := Db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT INTO Posts(user_id, post_text) VALUES(?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	result, err := stmt.Exec(userID, postText)
	if err != nil {
		return err
	}

	lastInsertID, err := result.LastInsertId()
	if err != nil {
		return err
	}

	for _, categoryName := range selectedCategories {
		var categoryID int
		err := tx.QueryRow("SELECT category_id FROM Categories WHERE category_name = ?", categoryName).Scan(&categoryID)
		if err != nil {
			return err
		}

		_, err = tx.Exec("INSERT INTO Post_Categories(post_id, category_id) VALUES(?, ?)", lastInsertID, categoryID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func getCategories() ([]Category, error) {
	rows, err := Db.Query("SELECT category_name FROM Categories")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var category Category
		err := rows.Scan(&category.CategoryName)
		if err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}
	return categories, nil
}
