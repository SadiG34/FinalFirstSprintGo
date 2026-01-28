package models

import (
	"sync"
	"time"
)

type User struct {
	ID       string    `json:"id"`
	Username string    `json:"username"`
	Created  time.Time `json:"created"`
}

type Post struct {
	ID      string    `json:"id"`
	Author  string    `json:"author"`
	Content string    `json:"content"`
	Likes   []string  `json:"likes"`
	Created time.Time `json:"created"`
}

type Storage struct {
	Users     map[string]*User
	Posts     []*Post
	UserIDSeq int64
	PostIDSeq int64
	Mu        sync.RWMutex
}