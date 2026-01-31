package models

type RegisterRequest struct {
	Username string `json:"username"`
}

type CreatePostRequest struct {
	Author  string `json:"author"`
	Content string `json:"content"`
}

type LikeRequest struct {
	PostID string `json:"post_id"`
	UserID string `json:"user_id"`
}