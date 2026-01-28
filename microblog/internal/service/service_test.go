package service

import (
	"testing"
	"time"
	"microblog/internal/models"
)

func TestService_Register(t *testing.T) {
	svc := NewService()

	tests := []struct {
		name      string
		username  string
		wantError bool
	}{
		{"valid registration", "user1", false},
		{"empty username", "", true},
		{"duplicate username", "user1", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := RegisterRequest{Username: tt.username}
			_, err := svc.Register(req)

			if tt.wantError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestService_CreatePost(t *testing.T) {
	svc := NewService()

	user, _ := svc.Register(RegisterRequest{Username: "author1"})

	tests := []struct {
		name      string
		author    string
		content   string
		wantError bool
	}{
		{"valid post", user.ID, "Hello world", false},
		{"empty author", "", "content", true},
		{"empty content", user.ID, "", true},
		{"invalid author", "invalid_user", "content", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := CreatePostRequest{
				Author:  tt.author,
				Content: tt.content,
			}
			_, err := svc.CreatePost(req)

			if tt.wantError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestService_GetPosts(t *testing.T) {
	svc := NewService()

	posts := svc.GetPosts()
	if len(posts) != 0 {
		t.Errorf("expected 0 posts, got %d", len(posts))
	}

	user, _ := svc.Register(RegisterRequest{Username: "user1"})
	svc.CreatePost(CreatePostRequest{
		Author:  user.ID,
		Content: "Post 1",
	})

	posts = svc.GetPosts()
	if len(posts) != 1 {
		t.Errorf("expected 1 post, got %d", len(posts))
	}
}

func TestService_LikePost(t *testing.T) {
	svc := NewService()

	user1, _ := svc.Register(RegisterRequest{Username: "user1"})
	user2, _ := svc.Register(RegisterRequest{Username: "user2"})

	post, _ := svc.CreatePost(CreatePostRequest{
		Author:  user1.ID,
		Content: "Test post",
	})

	tests := []struct {
		name      string
		postID    string
		userID    string
		wantError bool
	}{
		{"valid like", post.ID, user2.ID, false},
		{"duplicate like", post.ID, user2.ID, true},
		{"invalid post", "invalid_post", user2.ID, true},
		{"invalid user", post.ID, "invalid_user", true},
		{"empty post_id", "", user2.ID, true},
		{"empty user_id", post.ID, "", true},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := models.LikeRequest{ 
				PostID: tt.postID,
				UserID: tt.userID,
			}
			err := svc.LikePost(req)

			if tt.wantError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !tt.wantError {
				time.Sleep(10 * time.Millisecond) //

				posts := svc.GetPosts()
				found := false
				for _, p := range posts {
					if p.ID == post.ID {
						for _, like := range p.Likes {
							if like == tt.userID {
								found = true
								break
							}
						}
						break
					}
				}
				if !found {
					t.Errorf("like from %s was not added to post %s", tt.userID, post.ID)
				}
			}

			if i == 0 {
				time.Sleep(10 * time.Millisecond)
			}
		})
	}
}