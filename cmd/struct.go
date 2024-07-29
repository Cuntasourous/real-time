package forum

import (
	"database/sql"
	"time"
)

var Db *sql.DB // Declare db globally
var CookieName = "forum_session"
var SessionDuration = 24 * time.Hour // Session duration (24 hours)

type User struct {
	UserID      int       `json:"user_id"`
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	Password    string    `json:"password"`
	DateCreated time.Time `json:"date_created"`
}

type Post struct {
	PostID       int    `json:"post_id"`
	UserID       int    `json:"user_id"`
	PostText     string `json:"post_text"`
	PostDate     string `json:"post_date"`
	LikeCount    int    `json:"like_count"`
	DislikeCount int    `json:"dislike_count"`
	Username     string
	Categories   []string
}

type Comment struct {
	CommentID    int    `json:"comment_id"`
	PostID       int    `json:"post_id"`
	UserID       int    `json:"user_id"`
	CommentText  string `json:"comment_text"`
	CommentDate  string `json:"comment_date"`
	LikeCount    int    `json:"like_count"`
	DislikeCount int    `json:"dislike_count"`
	Username     string
	IsLiked      bool
	IsDisliked   bool
}

type Category struct {
	CategoryID   int    `json:"category_id"`
	CategoryName string `json:"category_name"`
}

type PopularCategory struct {
	CategoryID   int
	CategoryName string
	PostCount    int
}

type Session struct {
	ID        string    `json:"id"`
	UserID    int       `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

type PostLikes struct {
	UserID    int       `json:"user_id"`
	PostID    int       `json:"post_id"`
	CreatedAt time.Time `json:"created_at"`
}

type PostDislikes struct {
	UserID    int       `json:"user_id"`
	PostID    int       `json:"post_id"`
	CreatedAt time.Time `json:"created_at"`
}

type CommentLikes struct {
	UserID    int       `json:"user_id"`
	CommentID int       `json:"comment_id"`
	CreatedAt time.Time `json:"created_at"`
}

type CommentDislikes struct {
	UserID    int       `json:"user_id"`
	CommentID int       `json:"comment_id"`
	CreatedAt time.Time `json:"created_at"`
}

type UserProfile struct {
	Username       string
	Email          string
	DateCreated    time.Time
	Posts          []Post
	Comments       []Comment
	LikedPosts     []Post
	PostCount      int
	CommentCount   int
	LikedPostCount int
}
