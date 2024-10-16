package forum

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	// Get the session ID from the cookie
	sessionID, _ := getCookie(r, w, CookieName)
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
		IsLiked    bool
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
			IsLiked    bool
			IsDisliked bool
		}{
			Post:       post,
			IsLiked:    isLiked,
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

	OnlineUsers, err := getOnlineUsers()
	if err != nil {
		fmt.Print("Error getting online users")
		OnlineUsers = []OnlineUser{}
	}

	data := struct {
		LoggedInUser string
		Posts        []struct {
			Post
			IsLiked    bool
			IsDisliked bool
		}
		PopularCategory []PopularCategory
		OnlineUsers     []OnlineUser
	}{
		LoggedInUser:    username,
		Posts:           posts,
		PopularCategory: popularCategories,
		OnlineUsers:     OnlineUsers,
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
	if isAuthenticated(r, w) {
		log.Println("handleRoot: User authenticated, redirecting to /home")
		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return
	}
	log.Println("handleRoot: User not authenticated, redirecting to /login")
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func getOnlineUsers() ([]OnlineUser, error) {
	query := `
    SELECT user_id, username
    FROM users
    WHERE is_online = TRUE
    ORDER BY username
    `

	rows, err := Db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var onlineUsers []OnlineUser
	for rows.Next() {
		var user OnlineUser
		if err := rows.Scan(&user.UserID, &user.Username); err != nil {
			return nil, err
		}
		onlineUsers = append(onlineUsers, user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return onlineUsers, nil
}

func getAllUsers() ([]OnlineUser, error) {
	var users []OnlineUser
	rows, err := Db.Query("SELECT user_id, username FROM users") // Adjust your query as needed
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var user OnlineUser
		if err := rows.Scan(&user.UserID, &user.Username); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

func ChatHandler(w http.ResponseWriter, r *http.Request) {
	receiverID := r.URL.Query().Get("receiver_id")

	// Get the session ID from the cookie
	sessionID, _ := getCookie(r, w, CookieName)
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

	// OnlineUsers, err := getOnlineUsers()
	// if err != nil {
	// 	fmt.Print("Error getting online users")
	// 	OnlineUsers = []OnlineUser{}
	// }

	// receiverID := r.URL.Query().Get("receiver_id")
	// If receiverID is empty, try to get the last chatted user
	if receiverID == "" {
		lastReceiverID, err := getLastReceiverID(userID) // A function to fetch the last receiver ID
		if err != nil {
			log.Println("Error retrieving last receiver ID:", err)
			receiverID = "1" // Fallback to default user ID
		} else {
			receiverID = lastReceiverID
		}
	}

	// Get all users instead of online users
	allUsers, err := getAllUsers()
	if err != nil {
		log.Println("Error getting all users:", err)
		allUsers = []OnlineUser{}
	}

	messages, err := GetChatMessages(userID, receiverID)
	if err != nil {
		log.Println("Error retrieving messages:", err)
		messages = []ChatMessage{} // Initialize messages as empty slice
	}

	data := struct {
		LoggedInUser   string
		AllUsers       []OnlineUser
		Messages       []ChatMessage
		ReceiverID     string
		LoggedInUserID int
	}{
		LoggedInUser:   username,
		AllUsers:       allUsers,
		Messages:       messages,
		ReceiverID:     receiverID,
		LoggedInUserID: userID,
	}

	t, err := template.ParseFiles("templates/chat.html")
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

func ChatUpdatesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	sessionID, _ := getCookie(r, w, CookieName)
	var userID int
	err := Db.QueryRow("SELECT user_id FROM sessions WHERE id = ?", sessionID).Scan(&userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	receiverID := r.URL.Query().Get("receiver_id")

	for {
		messages, err := GetChatMessages(userID, receiverID)
		if err != nil {
			log.Println("Error retrieving messages:", err)
			return
		}

		data, err := json.Marshal(messages)
		if err != nil {
			log.Println("Error marshalling messages:", err)
			return
		}

		fmt.Fprintf(w, "data: %s\n\n", data)
		flusher.Flush()

		time.Sleep(2 * time.Second) // Poll every 2 seconds
	}
}

func SendMessageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Get the sender ID from the session
		sessionID, _ := getCookie(r, w, CookieName)
		var userID int
		err := Db.QueryRow("SELECT user_id FROM sessions WHERE id = ?", sessionID).Scan(&userID)
		if err != nil {
			http.Error(w, "Failed to get user ID", http.StatusInternalServerError)
			return
		}

		// Get the receiver ID from the query parameters
		receiverID := r.URL.Query().Get("receiver_id")
		message := r.FormValue("comment_text")

		// Insert the message into the Chats table
		_, err = Db.Exec("INSERT INTO Chats (sender_id, receiver_id, message) VALUES (?, ?, ?)", userID, receiverID, message)
		if err != nil {
			http.Error(w, "Failed to send message", http.StatusInternalServerError)
			return
		}

		// Redirect back to the chat page with the receiver ID
		http.Redirect(w, r, "/chat?receiver_id="+receiverID, http.StatusSeeOther)
	}
}

func getLastReceiverID(userID int) (string, error) {
	var lastReceiverID string
	err := Db.QueryRow(`
        SELECT CASE 
            WHEN sender_id = ? THEN receiver_id 
            ELSE sender_id 
        END AS last_receiver_id 
        FROM Chats 
        WHERE sender_id = ? OR receiver_id = ? 
        ORDER BY created_at DESC 
        LIMIT 1`, userID, userID, userID).Scan(&lastReceiverID)
	return lastReceiverID, err
}

func GetChatMessages(userID int, receiverID string) ([]ChatMessage, error) {
	var messages []ChatMessage
	rows, err := Db.Query(`
        SELECT c.sender_id, c.receiver_id, c.message, c.created_at, u.username
        FROM Chats c
        JOIN users u ON c.sender_id = u.user_id
        WHERE (c.sender_id = ? AND c.receiver_id = ?) OR (c.sender_id = ? AND c.receiver_id = ?) 
        ORDER BY c.created_at ASC`, userID, receiverID, receiverID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var msg ChatMessage
		if err := rows.Scan(&msg.SenderID, &msg.ReceiverID, &msg.Message, &msg.CreatedAt, &msg.SenderName); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	return messages, nil
}
