package handlers

import (
	"encoding/json"
	"microblog/internal/service"
	"net/http"

	"github.com/gorilla/mux"
)

type Handlers struct {
	service *service.Service
}

func NewHandlers(s *service.Service) *Handlers {
	return &Handlers{service: s}
}

func (h *Handlers) Register(w http.ResponseWriter, r *http.Request) {
	var req service.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := h.service.Register(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *Handlers) CreatePost(w http.ResponseWriter, r *http.Request) {
	var req service.CreatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := h.service.CreatePost(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *Handlers) GetPosts(w http.ResponseWriter, r *http.Request) {
	posts := h.service.GetPosts()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

func (h *Handlers) LikePost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postID := vars["id"]

	var req struct {
		UserID string `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	likeReq := service.LikeRequest{
		PostID: postID,
		UserID: req.UserID,
	}

	if err := h.service.LikePost(likeReq); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handlers) SetupRoutes(r *mux.Router) {
	r.HandleFunc("/register", h.Register).Methods("POST")
	r.HandleFunc("/posts", h.CreatePost).Methods("POST")
	r.HandleFunc("/posts", h.GetPosts).Methods("GET")
	r.HandleFunc("/posts/{id}/like", h.LikePost).Methods("POST")
}