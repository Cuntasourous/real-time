package forum

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func CategoryHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// Handle GET request
		renderCategoryForm(w, r)
	case http.MethodPost:
		// Handle POST request
		handleCreateCategory(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func ViewCategoriesHandler(w http.ResponseWriter, r *http.Request) {
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
	var categories []Category

	rows, err := Db.Query("SELECT category_id, category_name FROM Categories")
	if err != nil {
		http.Error(w, "Error fetching categories", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var category Category
		if err := rows.Scan(&category.CategoryID, &category.CategoryName); err != nil {
			http.Error(w, "Error scanning categories", http.StatusInternalServerError)
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

	t, err := template.ParseFiles("templates/view_categories.html")
	if err != nil {
		http.Error(w, "Error parsing template", http.StatusInternalServerError)
		return
	}

	err = t.Execute(w, data)
	if err != nil {
		// Log the error instead of sending it to the client, as headers have already been written
		log.Printf("Error executing template: %v", err)
	}
}

func ViewCategoryPostsHandler(w http.ResponseWriter, r *http.Request) {
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

	// Extract the category ID from the URL
	path := strings.TrimPrefix(r.URL.Path, "/category/")
	categoryID, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	// Fetch the category name
	var categoryName string
	err = Db.QueryRow("SELECT category_name FROM Categories WHERE category_id = ?", categoryID).Scan(&categoryName)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Category not found", http.StatusNotFound)
		} else {
			http.Error(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	// Fetch all posts for this category
	rows, err := Db.Query(`
        SELECT p.post_id, p.user_id, p.post_text, p.post_date, p.like_count, p.dislike_count, u.username
        FROM Posts p
        JOIN Post_Categories pc ON p.post_id = pc.post_id
        JOIN Users u ON p.user_id = u.user_id
        WHERE pc.category_id = ?
    `, categoryID)
	if err != nil {
		http.Error(w, "Error fetching posts", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		err := rows.Scan(&post.PostID, &post.UserID, &post.PostText, &post.PostDate, &post.LikeCount, &post.DislikeCount, &post.Username)
		if err != nil {
			http.Error(w, "Error scanning posts", http.StatusInternalServerError)
			return
		}
		posts = append(posts, post)
	}

	// Prepare data for the template
	data := struct {
		CategoryName string
		Posts        []Post
		LoggedInUser string
	}{
		CategoryName: categoryName,
		Posts:        posts,
		LoggedInUser: username,
	}

	// Parse and execute the template
	t, err := template.ParseFiles("templates/category_posts.html")
	if err != nil {
		http.Error(w, "Error parsing template", http.StatusInternalServerError)
		return
	}

	err = t.Execute(w, data)
	if err != nil {
		log.Printf("Error executing template: %v", err)
	}
}

func renderCategoryForm(w http.ResponseWriter, r *http.Request) {
	// log.Println("Rendering category creation form")

	sessionID, _ := getCookie(r, CookieName)
	var userID int
	err := Db.QueryRow("SELECT user_id FROM sessions WHERE id = ?", sessionID).Scan(&userID)
	if err != nil {
		fmt.Println("guest")
	}

	var username string
	err = Db.QueryRow("SELECT username FROM users WHERE user_id = ?", userID).Scan(&username)
	if err != nil {
		username = ""
	}

	popularCategories, err := getPopularCategories()
	if err != nil {
		log.Printf("Error fetching popular categories: %v", err)
		// Instead of handling the error here, we'll pass an empty slice
		popularCategories = []PopularCategory{}
	}

	// Create a struct to hold both the logged-in username and the users slice
	data := struct {
		LoggedInUser    string
		PopularCategory []PopularCategory
	}{
		LoggedInUser:    username,
		PopularCategory: popularCategories,
	}

	t, err := template.ParseFiles("templates/create_category.html")
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = t.Execute(w, data)
	if err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// New function to get categories for a post
func getCategoriesByPostID(postID int) ([]string, error) {
	rows, err := Db.Query(`
		SELECT c.category_name 
		FROM Categories c
		JOIN Post_Categories pc ON c.category_id = pc.category_id
		WHERE pc.post_id = ?
	`, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []string
	for rows.Next() {
		var category string
		if err := rows.Scan(&category); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	return categories, nil
}

func handleCreateCategory(w http.ResponseWriter, r *http.Request) {
	log.Println("Processing POST request for category creation")

	categoryName := r.FormValue("category_name")
	if categoryName == "" {
		log.Println("Empty category name submitted")
		http.Error(w, "Please provide a category name", http.StatusBadRequest)
		return
	}

	log.Printf("Attempting to create category: %s", categoryName)

	// Start a transaction
	tx, err := Db.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback() // Roll back the transaction if it's not committed

	// Insert the category into the Categories table
	result, err := tx.Exec("INSERT INTO Categories(category_name) VALUES(?)", categoryName)
	if err != nil {
		log.Printf("Error inserting category: %v", err)
		http.Error(w, "Error creating category", http.StatusInternalServerError)
		return
	}

	// Get the last inserted ID
	lastInsertID, err := result.LastInsertId()
	if err != nil {
		log.Printf("Error getting last inserted ID: %v", err)
	} else {
		log.Printf("Category created with ID: %d", lastInsertID)
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		http.Error(w, "Error creating category", http.StatusInternalServerError)
		return
	}

	log.Println("Category created successfully")
	http.Redirect(w, r, "/home", http.StatusSeeOther)
}

func getPopularCategories() ([]PopularCategory, error) {
	query := `
    SELECT c.category_id, c.category_name, COUNT(pc.post_id) as post_count
    FROM Categories c
    LEFT JOIN Post_Categories pc ON c.category_id = pc.category_id
    GROUP BY c.category_id, c.category_name
    ORDER BY post_count DESC
    LIMIT 5
    `

	rows, err := Db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []PopularCategory
	for rows.Next() {
		var cat PopularCategory
		if err := rows.Scan(&cat.CategoryID, &cat.CategoryName, &cat.PostCount); err != nil {
			return nil, err
		}
		categories = append(categories, cat)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}
