package forum

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
)

func Handler(w http.ResponseWriter, r *http.Request) {

	// Get the session ID from the cookie
	sessionID, _ := getCookie(r,w, CookieName)
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
		ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var posts []struct {
		Post
		IsLiked bool
		IsDisliked bool
	}

	for rows.Next() {
		var post Post
		var categoriesString string
		err := rows.Scan(&post.PostID, &post.UserID, &post.PostText, &post.PostDate, &post.LikeCount, &post.DislikeCount, &post.Username, &categoriesString)
		if err != nil {
			ErrorHandler(w, r, http.StatusInternalServerError)
			return
		}
		post.Categories = strings.Split(categoriesString, ",")

		// Check if the current user has liked the post
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

		posts = append(posts, struct {
			Post
			IsLiked bool
			IsDisliked bool
		}{
			Post:    post,
			IsLiked: isLiked,
			IsDisliked: isDisliked,
		})
	}

	if err = rows.Err(); err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError)
		fmt.Println("Error iterating over database results")
		return
	}

	popularCategories, err := getPopularCategories()
	if err != nil {
		log.Printf("Error fetching popular categories: %v", err)
		popularCategories = []PopularCategory{}
	}

	data := struct {
		LoggedInUser    string
		Posts           []struct {
			Post
			IsLiked bool
			IsDisliked bool
		}
		PopularCategory []PopularCategory
	}{
		LoggedInUser:    username,
		Posts:           posts,
		PopularCategory: popularCategories,
	}

	t, err := template.ParseFiles("templates/index.html")
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}
	err = t.Execute(w, data)
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError)
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
