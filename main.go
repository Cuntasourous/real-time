package main

import (
	"database/sql"
	forum "forum/cmd"
	"log"
	"net/http"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	var err error
	forum.Db, err = sql.Open("sqlite3", "test_forum.db")
	if err != nil {
		log.Fatal(err)
	}
	defer forum.Db.Close()

	// Test the connection
	err = forum.Db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	_, err = forum.Db.Exec("PRAGMA journal_mode=WAL;")
	if err != nil {
		log.Fatalf("Error setting journal_mode: %v", err)
	}

	http.HandleFunc("/like/", forum.LikeHandler)
	http.HandleFunc("/dislike/", forum.DislikeHandler)
	http.HandleFunc("/clike/", forum.CommentikeHandler)
	http.HandleFunc("/cdislike/", forum.CommentDislikeHandler)
	// http.HandleFunc("/like2/", forum.LikeHandler2)
	// http.HandleFunc("/dislike2/", forum.DislikeHandler2)
	http.HandleFunc("/home", forum.Handler)
	http.HandleFunc("/", forum.HandleRoot)
	http.HandleFunc("/debug", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Debug: Request to %s", r.URL.Path)
	})
	http.HandleFunc("/logout", forum.HandleLogout)
	http.HandleFunc("/register", forum.RegisterHandler)
	http.HandleFunc("/login", forum.HandleLogin)
	http.HandleFunc("/create_post", forum.PostHandler)
	http.HandleFunc("/create_category/", forum.CategoryHandler)
	http.HandleFunc("/view_categories/", forum.ViewCategoriesHandler)
	http.HandleFunc("/category/", forum.ViewCategoryPostsHandler)
	http.HandleFunc("/view_post/", forum.HandleViewPost)
	http.HandleFunc("/add_comment/", forum.HandleAddCommentAJAX)
	http.HandleFunc("/profile", forum.ViewProfileHandler)
	//Error pages
	// Serve static files
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	js := http.FileServer(http.Dir("js"))
	http.Handle("/js/", http.StripPrefix("/js/", js))

	//go live
	log.Print("http://localhost:8080/")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("ListenAndServe: %v", err)
	}
}
