package service

import (
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"microblog/internal/logger"
	"microblog/internal/models"
	"microblog/internal/queue"
)

type Service struct {
	storage     *models.Storage
	likeQueue   *queue.LikeQueue
	eventLogger *logger.EventLogger
}

type ServiceInterface interface {
	Register(req models.RegisterRequest) (*RegisterResponse, error)
	CreatePost(req models.CreatePostRequest) (*CreatePostResponse, error)
	GetPosts() []PostResponse
	LikePost(req models.LikeRequest) error
	Close()
}

func NewService() *Service {
	s := &Service{
		storage: &models.Storage{
			Users: make(map[string]*models.User),
			Posts: make([]*models.Post, 0),
		},
		eventLogger: logger.NewEventLogger(),
	}
	s.likeQueue = queue.NewLikeQueue(s.processLike)
	return s
}

func (s *Service) Close() {
	s.likeQueue.Close()
	s.eventLogger.Close()
}

func (s *Service) processLike(req models.LikeRequest) error {
	s.storage.Mu.Lock()
	defer s.storage.Mu.Unlock()

	for _, post := range s.storage.Posts {
		if post.ID == req.PostID {
			post.Likes = append(post.Likes, req.UserID)
			s.eventLogger.Log("LIKE", fmt.Sprintf("User %s liked post %s", req.UserID, req.PostID))
			return nil
		}
	}
	return errors.New("post not found")
}

type RegisterResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

func (s *Service) Register(req models.RegisterRequest) (*RegisterResponse, error) {
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
	s.eventLogger.Log("REGISTER", fmt.Sprintf("User %s created with ID %s", req.Username, user.ID))

	return &RegisterResponse{
		ID:       user.ID,
		Username: user.Username,
	}, nil
}

type CreatePostResponse struct {
	ID      string `json:"id"`
	Author  string `json:"author"`
	Content string `json:"content"`
}

func (s *Service) CreatePost(req models.CreatePostRequest) (*CreatePostResponse, error) {
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

	s.eventLogger.Log("POST_CREATED", fmt.Sprintf("Post %s by %s", post.ID, req.Author))

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

func (s *Service) LikePost(req models.LikeRequest) error {
	if req.PostID == "" || req.UserID == "" {
		return errors.New("post_id and user_id are required")
	}

	s.storage.Mu.RLock()
	userExists := s.storage.Users[req.UserID] != nil
	postExists := false
	alreadyLiked := false

	for _, post := range s.storage.Posts {
		if post.ID == req.PostID {
			postExists = true
			for _, like := range post.Likes {
				if like == req.UserID {
					alreadyLiked = true
					break
				}
			}
			break
		}
	}
	s.storage.Mu.RUnlock()

	if !userExists {
		return errors.New("user not found")
	}
	if !postExists {
		return errors.New("post not found")
	}
	if alreadyLiked {
		return errors.New("post already liked")
	}

	s.likeQueue.Add(req)
	return nil
}