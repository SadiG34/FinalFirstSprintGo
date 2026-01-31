package queue

import (
	"log"
	"microblog/internal/models"
)

type LikeProcessor func(req models.LikeRequest) error

type LikeQueue struct {
	ch        chan models.LikeRequest
	processor LikeProcessor
}

func NewLikeQueue(processor LikeProcessor) *LikeQueue {
	q := &LikeQueue{
		ch:        make(chan models.LikeRequest, 100),
		processor: processor,
	}
	go q.worker()
	return q
}

func (q *LikeQueue) Add(req models.LikeRequest) {
	q.ch <- req
}

func (q *LikeQueue) Close() {
	close(q.ch)
}

func (q *LikeQueue) worker() {
	for req := range q.ch {
		if err := q.processor(req); err != nil {
			log.Printf("Failed to process like: %v", err)
		}
	}
}