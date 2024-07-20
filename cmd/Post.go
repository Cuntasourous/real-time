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
	}

	var username string
	err = Db.QueryRow("SELECT username FROM users WHERE user_id = ?", userID).Scan(&username)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}

	if r.Method == "POST" {
		fmt.Println("post")
		post_text := r.FormValue("post_text")
		// Get selected categories
		selectedCategories := r.Form["categories"] // Get all selected categories

		if post_text == "" {
			http.Error(w, "Please add some text", http.StatusBadRequest)
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

		// Insert the post into the Posts table
		stmt, err := tx.Prepare("INSERT INTO Posts(user_id, post_text) VALUES(?, ?)")
		if err != nil {
			log.Printf("Error preparing SQL statement: %v", err)
			http.Error(w, "Error preparing SQL statement", http.StatusInternalServerError)
			return
		}
		defer stmt.Close()

		result, err := stmt.Exec(userID, post_text)
		if err != nil {
			log.Printf("Error inserting post: %v", err)
			http.Error(w, "Error inserting post", http.StatusInternalServerError)
			return
		}

		// Get the last inserted post ID
		lastInsertID, err := result.LastInsertId()
		if err != nil {
			log.Printf("Error getting last inserted ID: %v", err)
			http.Error(w, "Error getting last inserted ID", http.StatusInternalServerError)
			return
		}

		// Insert the post-category associations into the Post_Categories table
		for _, categoryName := range selectedCategories {
			var categoryID int
			err := Db.QueryRow("SELECT category_id FROM Categories WHERE category_name = ?", categoryName).Scan(&categoryID)
			if err != nil {
				log.Printf("Error getting category ID: %v", err)
				http.Error(w, "Error getting category ID", http.StatusInternalServerError)
				return
			}

			_, err = tx.Exec("INSERT INTO Post_Categories(post_id, category_id) VALUES(?, ?)", lastInsertID, categoryID)
			if err != nil {
				log.Printf("Error inserting post-category association: %v", err)
				http.Error(w, "Error inserting post-category association", http.StatusInternalServerError)
				return
			}
		}

		// Commit the transaction
		err = tx.Commit()
		if err != nil {
			log.Printf("Error committing transaction: %v", err)
			http.Error(w, "Error committing transaction", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/home", http.StatusSeeOther)
	} else {
		// Get the list of categories from the database
		rows, err := Db.Query("SELECT category_name FROM Categories")
		if err != nil {
			log.Printf("Error getting categories: %v", err)
			http.Error(w, "Error getting categories", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var categories []Category
		for rows.Next() {
			var category Category
			err := rows.Scan(&category.CategoryName)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			categories = append(categories, category)
		}

		// Pass the categories to the template
		data := struct {
			LoggedInUser string
			Categories   []Category
		}{
			LoggedInUser: username,
			Categories:   categories,
		}

		// Render the create_post template
		t, err := template.ParseFiles("templates/create_post.html")
		if err != nil {
			http.Error(w, "Error parsing template", http.StatusInternalServerError)
			return
		}
		err = t.Execute(w, data)
		if err != nil {
			http.Error(w, "Error executing template", http.StatusInternalServerError)
			return
		}
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

func HandleViewPost(w http.ResponseWriter, r *http.Request) {
	// Extract the post_id from the URL
	postID, err := getPostIDFromURL(r)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	sessionID, _ := getCookie(r, CookieName)
	var userID int
	err = Db.QueryRow("SELECT user_id FROM sessions WHERE id = ?", sessionID).Scan(&userID)
	if err != nil {
		fmt.Println("guest")
	}

	var username string
	err = Db.QueryRow("SELECT username FROM users WHERE user_id = ?", userID).Scan(&username)
	if err != nil {
		username = ""
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

	// Fetch the post data from the database using postID
	post, err := getPostByID(postID)
	if err != nil {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	// Fetch categories for this post
	categories, err := getCategoriesByPostID(postID)
	if err != nil {
		http.Error(w, "Error fetching categories", http.StatusInternalServerError)
		return
	}

	// Fetch comments for the post
	comments, err := getCommentsByPostID(postID)
	if err != nil {
		http.Error(w, "Error fetching comments", http.StatusInternalServerError)
		return
	}

	if r.Method == http.MethodPost {
		// Handle POST request
		handleAddComment(w, r, postID)
	}

	// Fetch popular categories
	popularCategories, err := getPopularCategories()
	if err != nil {
		log.Printf("Error fetching popular categories: %v", err)
		// Instead of handling the error here, we'll pass an empty slice
		popularCategories = []PopularCategory{}
	}

	// Render the view_post template
	data := map[string]interface{}{
		"Post":            post,
		"Categories":      categories,
		"Comments":        comments,
		"LoggedInUser":    username,
		"PopularCategory": popularCategories,
	}
	// Parse the template file
	t, err := template.ParseFiles("templates/view_post.html")
	if err != nil {
		http.Error(w, "Error parsing template", http.StatusInternalServerError)
		return
	}

	// Execute the template with the data
	err = t.Execute(w, data)
	if err != nil {
		http.Error(w, "Error executing template", http.StatusInternalServerError)
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
