package store

import (
	"fmt"
	"log/slog"
	"slices"
	"sync"

	"github.com/bpietroniro/requestjar-go/internal/errors"
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
	slog.Info("creating request storage dependency")
	return &requestStore{requests: make(map[string][]*models.Request)}
}

func (s *requestStore) CreateRequest(jarID string, req *models.Request) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	requests, jarExists := s.requests[jarID]

	if !jarExists {
		return errors.NotFound("jar not found")
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
		s.requests[jarID] = make([]*models.Request, 0, 5)
	} else {
		slog.Warn(fmt.Sprintf("Jar %s already existed in request store", jarID))
	}

	return nil
}

func (s *requestStore) List(jarID string) ([]*models.Request, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	requests, jarExists := s.requests[jarID]

	if !jarExists {
		return nil, errors.NotFound("jar not found")
	}

	return requests, nil
}

func (s *requestStore) DeleteOneRequest(jarID string, reqID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	requests, jarExists := s.requests[jarID]

	if !jarExists {
		return errors.NotFound("jar not found")
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

	_, jarExists := s.requests[jarID]

	if !jarExists {
		return errors.NotFound("no requests record found for jar")
	}

	delete(s.requests, jarID)
	return nil
}
