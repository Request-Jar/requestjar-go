package store

import (
	"errors"
	"log"
	"slices"
	"sync"

	"github.com/bpietroniro/requestjar-go/internal/models"
)

type RequestStore interface {
	CreateRequest(jarID string, req *models.Request) error
	CreateJarKey(jarID string) error
	List(jarID string) ([]*models.Request, error)
	DeleteOneRequest(jarID string, reqID string) error
	DeleteAllRrequests(jarID string) error
}

type requestStore struct {
	requests map[string][]*models.Request
	mu       sync.RWMutex
}

func NewInMemoryRequestStore() RequestStore {
	return &requestStore{requests: make(map[string][]*models.Request)}
}

func (s *requestStore) CreateRequest(jarID string, req *models.Request) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	requests, jarExists := s.requests[jarID]

	if !jarExists {
		log.Println("jar not found")
		return errors.New("jar not found")
	} else {
		s.requests[jarID] = append(requests, req)
	}

	return nil
}

func (s *requestStore) CreateJarKey(jarID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, jarExists := s.requests[jarID]

	if !jarExists {
		log.Printf("Jar %s didn't exist, creating new request slice", jarID)
		s.requests[jarID] = make([]*models.Request, 0, 5)
	} else {
		log.Printf("Jar %s already existed in request store", jarID)
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

func (s *requestStore) DeleteOneRequest(jarID string, reqID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	requests, jarExists := s.requests[jarID]

	if !jarExists {
		return errors.New("jar not found")
	}

	filteredRequests := slices.DeleteFunc(requests, func(r *models.Request) bool {
		return r.ID == reqID
	})

	s.requests[jarID] = filteredRequests
	return nil
}

func (s *requestStore) DeleteAllRrequests(jarID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.requests, jarID)
	return nil
}
