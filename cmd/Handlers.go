package forum

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	// if !isAuthenticated(r) {
	// 	http.Redirect(w, r, "/login", http.StatusSeeOther)
	// 	fmt.Println("inauthenticated so we sent you to login")
	// 	return
	// }

	// Get the session ID from the cookie
	sessionID, _ := getCookie(r, CookieName)
	var userID int
	err := Db.QueryRow("SELECT user_id FROM sessions WHERE id = ?", sessionID).Scan(&userID)
	if err != nil {
		// http.Redirect(w, r, "/login", http.StatusSeeOther)
		// return
		fmt.Println("guest")
	}

	var username string
	err = Db.QueryRow("SELECT username FROM users WHERE user_id = ?", userID).Scan(&username)
	if err != nil {
		// http.Redirect(w, r, "/login", http.StatusSeeOther)
		username = ""
		// return
	}

	// Query the database for all posts
	rows, err := Db.Query(`
        SELECT 
            p.post_id, 
            p.user_id, 
            p.post_text, 
            p.post_date, 
            p.like_count, 
            p.dislike_count, 
            u.username, 
            GROUP_CONCAT(c.category_name) AS categories 
        FROM Posts p
        JOIN Users u ON p.user_id = u.user_id
        JOIN Post_Categories pc ON p.post_id = pc.post_id
        JOIN Categories c ON pc.category_id = c.category_id
        GROUP BY p.post_id
        ORDER BY p.post_date DESC
    `)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		var categoriesString string                                                                                                                          // Declare a variable to hold the categories string
		err := rows.Scan(&post.PostID, &post.UserID, &post.PostText, &post.PostDate, &post.LikeCount, &post.DislikeCount, &post.Username, &categoriesString) // Scan the categories string
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Split the categories string into a slice
		post.Categories = strings.Split(categoriesString, ",") // Split the categories string
		posts = append(posts, post)
	}

	// Check for errors during iteration
	if err = rows.Err(); err != nil {
		http.Error(w, "Error iterating over database results", http.StatusInternalServerError)
		return
	}

	// Fetch popular categories
	popularCategories, err := getPopularCategories()
	if err != nil {
		log.Printf("Error fetching popular categories: %v", err)
		// Instead of handling the error here, we'll pass an empty slice
		popularCategories = []PopularCategory{}
	}

	// Create a struct to hold both the logged-in username and the users slice
	data := struct {
		LoggedInUser    string
		Posts           []Post
		PopularCategory []PopularCategory
	}{
		LoggedInUser:    username,
		Posts:           posts,
		PopularCategory: popularCategories,
	}

	// Render the index template
	t, err := template.ParseFiles("templates/index.html")
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

func HandleRoot(w http.ResponseWriter, r *http.Request) {
	log.Printf("handleRoot: Request to %s", r.URL.Path)
	if r.URL.Path != "/" {
		ErrorHandler(w, r, http.StatusNotFound)
		return
	}
	if isAuthenticated(r) {
		log.Println("handleRoot: User authenticated, redirecting to /home")
		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return
	}
	log.Println("handleRoot: User not authenticated, redirecting to /login")
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
