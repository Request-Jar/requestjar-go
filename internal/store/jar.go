package store

import (
	"errors"
	"sync"
	"time"

	"github.com/bpietroniro/requestjar-go/internal/models"
	"github.com/bpietroniro/requestjar-go/internal/util"
)

type JarStore interface {
	Create(name string) (string, error)
	Get(id string) (*models.Jar, error)
	List() ([]*models.Jar, error)
	Delete(id string) error
}

type jarStore struct {
	jars map[string]*models.Jar
	mu   sync.RWMutex
}

func NewInMemoryJarStore() JarStore {
	return &jarStore{jars: make(map[string]*models.Jar)}
}

func (s *jarStore) Create(name string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := util.GenerateID()
	s.jars[id] = &models.Jar{ID: id, Name: name, CreatedAt: time.Now()}

	return id, nil
}

func (s *jarStore) Get(id string) (*models.Jar, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	jar, exists := s.jars[id]
	if !exists {
		return nil, errors.New("jar not found")
	}

	return jar, nil
}

func (s *jarStore) List() ([]*models.Jar, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	jars := []*models.Jar{}

	for _, j := range s.jars {
		jars = append(jars, j)
	}

	return jars, nil
}

func (s *jarStore) Delete(jarID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.jars, jarID)
	return nil
}
