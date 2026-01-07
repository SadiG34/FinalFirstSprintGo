package service

import (
	"errors"
	"fmt"
	"microblog/internal/models"
	"sync/atomic"
	"time"
)

type Service struct {
	storage *models.Storage
}

func NewService() *Service {
	return &Service{
		storage: &models.Storage{
			Users: make(map[string]*models.User),
			Posts: make([]*models.Post, 0),
		},
	}
}

type RegisterRequest struct {
	Username string `json:"username"`
}

type RegisterResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

func (s *Service) Register(req RegisterRequest) (*RegisterResponse, error) {
	if req.Username == "" {
		return nil, errors.New("username is required")
	}

	s.storage.Mu.Lock()
	defer s.storage.Mu.Unlock()

	for _, user := range s.storage.Users {
		if user.Username == req.Username {
			return nil, errors.New("username already exists")
		}
	}

	userID := atomic.AddInt64(&s.storage.UserIDSeq, 1)
	user := &models.User{
		ID:       fmt.Sprintf("user_%d", userID),
		Username: req.Username,
		Created:  time.Now(),
	}

	s.storage.Users[user.ID] = user

	return &RegisterResponse{
		ID:       user.ID,
		Username: user.Username,
	}, nil
}

type CreatePostRequest struct {
	Author  string `json:"author"`
	Content string `json:"content"`
}

type CreatePostResponse struct {
	ID      string `json:"id"`
	Author  string `json:"author"`
	Content string `json:"content"`
}

func (s *Service) CreatePost(req CreatePostRequest) (*CreatePostResponse, error) {
	if req.Author == "" {
		return nil, errors.New("author is required")
	}
	if req.Content == "" {
		return nil, errors.New("content is required")
	}

	s.storage.Mu.RLock()
	_, exists := s.storage.Users[req.Author]
	s.storage.Mu.RUnlock()

	if !exists {
		return nil, errors.New("author not found")
	}

	postID := atomic.AddInt64(&s.storage.PostIDSeq, 1)
	post := &models.Post{
		ID:      fmt.Sprintf("post_%d", postID),
		Author:  req.Author,
		Content: req.Content,
		Likes:   make([]string, 0),
		Created: time.Now(),
	}

	s.storage.Mu.Lock()
	s.storage.Posts = append(s.storage.Posts, post)
	s.storage.Mu.Unlock()

	return &CreatePostResponse{
		ID:      post.ID,
		Author:  post.Author,
		Content: post.Content,
	}, nil
}

type PostResponse struct {
	ID      string    `json:"id"`
	Author  string    `json:"author"`
	Content string    `json:"content"`
	Likes   []string  `json:"likes"`
	Created time.Time `json:"created"`
}

func (s *Service) GetPosts() []PostResponse {
	s.storage.Mu.RLock()
	defer s.storage.Mu.RUnlock()

	posts := make([]PostResponse, len(s.storage.Posts))
	for i, post := range s.storage.Posts {
		posts[i] = PostResponse{
			ID:      post.ID,
			Author:  post.Author,
			Content: post.Content,
			Likes:   post.Likes,
			Created: post.Created,
		}
	}

	return posts
}

type LikeRequest struct {
	PostID string `json:"post_id"`
	UserID string `json:"user_id"`
}

func (s *Service) LikePost(req LikeRequest) error {
	if req.PostID == "" {
		return errors.New("post_id is required")
	}
	if req.UserID == "" {
		return errors.New("user_id is required")
	}

	s.storage.Mu.RLock()
	_, userExists := s.storage.Users[req.UserID]
	s.storage.Mu.RUnlock()

	if !userExists {
		return errors.New("user not found")
	}

	s.storage.Mu.Lock()
	defer s.storage.Mu.Unlock()

	for _, post := range s.storage.Posts {
		if post.ID == req.PostID {
			for _, like := range post.Likes {
				if like == req.UserID {
					return errors.New("post already liked")
				}
			}
			post.Likes = append(post.Likes, req.UserID)
			return nil
		}
	}

	return errors.New("post not found")
}