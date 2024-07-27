package forum

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func PostHandler(w http.ResponseWriter, r *http.Request) {
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
		handleAddComment(w, r, postID)
	}

	// Get userID from the session
	sessionID, _ := getCookie(r, CookieName)
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

	// Handle like and dislike actions
	if r.Method == http.MethodPost {
		action := r.URL.Path
		if strings.HasPrefix(action, "/like2/") {
			LikeHandler(w, r)
			return
		} else if strings.HasPrefix(action, "/dislike2/") {
			DislikeHandler(w, r)
			return
		}
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
