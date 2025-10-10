package store

import (
	"errors"
	"sync"

	"github.com/bpietroniro/requestjar-go/internal/models"
)

type RequestStore interface {
	Create(jarID string, req *models.Request) error
	List(jarID string) ([]*models.Request, error)
}

type requestStore struct {
	requests map[string][]*models.Request
	mu       sync.RWMutex
}

func NewInMemoryRequestStore(jarID string) RequestStore {
	return &requestStore{requests: make(map[string][]*models.Request)}
}

func (s *requestStore) Create(jarID string, req *models.Request) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	requests, jarExists := s.requests[jarID]

	if !jarExists {
		s.requests[jarID] = []*models.Request{req}
	} else {
		requests = append(requests, req)
		s.requests[jarID] = requests
	}

	return nil
}

func (s *requestStore) List(jarID string) ([]*models.Request, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	requests, jarExists := s.requests[jarID]

	if !jarExists {
		return nil, errors.New("jar not found")
	}

	return requests, nil
}
