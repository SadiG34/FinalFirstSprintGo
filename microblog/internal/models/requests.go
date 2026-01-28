package models

type LikeRequest struct {
	PostID string `json:"post_id"`
	UserID string `json:"user_id"`
}