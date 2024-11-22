package forum

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
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

func getAllUsers(currentUserID int) ([]OnlineUser, error) {
	query := `
        WITH LastMessages AS (
            SELECT 
                CASE 
                    WHEN sender_id = ? THEN receiver_id
                    WHEN receiver_id = ? THEN sender_id
                END as other_user_id,
                MAX(created_at) as last_message_time
            FROM Chats
            WHERE sender_id = ? OR receiver_id = ?
            GROUP BY other_user_id
        )
        SELECT 
            u.user_id,
            u.username,
            u.is_online,  -- Include the is_online field
            lm.last_message_time
        FROM users u
        LEFT JOIN LastMessages lm ON u.user_id = lm.other_user_id
        WHERE u.user_id != ?
        ORDER BY 
            lm.last_message_time DESC NULLS LAST,
            u.username ASC
    `

	rows, err := Db.Query(query, currentUserID, currentUserID, currentUserID, currentUserID, currentUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []OnlineUser
	for rows.Next() {
		var user OnlineUser
		var lastMessageTime sql.NullString
		if err := rows.Scan(&user.UserID, &user.Username, &user.IsOnline, &lastMessageTime); err != nil {
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
	allUsers, err := getAllUsers(userID)
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

func ChatHistoryHandler(w http.ResponseWriter, r *http.Request) {
	receiverID := r.URL.Query().Get("receiver_id")
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	rows, err := Db.Query(`
        SELECT id, sender_id, receiver_id, message, created_at, 
               (SELECT username FROM users WHERE user_id = sender_id) as sender_name
        FROM Chats
        WHERE (sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)
        ORDER BY created_at DESC
        LIMIT ? OFFSET ?`, receiverID, receiverID, receiverID, receiverID, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var messages []ChatMessage
	for rows.Next() {
		var msg ChatMessage
		if err := rows.Scan(&msg.ID, &msg.SenderID, &msg.ReceiverID, &msg.Message, &msg.CreatedAt, &msg.SenderName); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		messages = append(messages, msg)
	}

	// Reverse the messages to display in chronological order (as the frontend prepends them)
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	json.NewEncoder(w).Encode(messages)
}

func NewMessagesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")

	receiverID := r.URL.Query().Get("receiver_id")
	since := r.URL.Query().Get("since")

	sessionID, _ := getCookie(r, w, CookieName)
	var userID int
	if err := Db.QueryRow("SELECT user_id FROM sessions WHERE id = ?", sessionID).Scan(&userID); err != nil {
		json.NewEncoder(w).Encode([]ChatMessage{})
		return
	}

	// Use a prepared statement for better performance
	stmt, err := Db.Prepare(`
        SELECT c.sender_id, c.receiver_id, c.message, c.created_at, u.username
        FROM Chats c
        JOIN users u ON c.sender_id = u.user_id
        WHERE ((c.sender_id = ? AND c.receiver_id = ?) OR (c.sender_id = ? AND c.receiver_id = ?))
        AND c.created_at > ?
        ORDER BY c.created_at ASC
        LIMIT 50`) // Add a limit to prevent excessive data transfer
	if err != nil {
		json.NewEncoder(w).Encode([]ChatMessage{})
		return
	}
	defer stmt.Close()

	rows, err := stmt.Query(userID, receiverID, receiverID, userID, since)
	if err != nil {
		json.NewEncoder(w).Encode([]ChatMessage{})
		return
	}
	defer rows.Close()

	messages := make([]ChatMessage, 0)
	for rows.Next() {
		var msg ChatMessage
		if err := rows.Scan(&msg.SenderID, &msg.ReceiverID, &msg.Message, &msg.CreatedAt, &msg.SenderName); err != nil {
			continue
		}
		messages = append(messages, msg)
	}

	json.NewEncoder(w).Encode(messages)
}
