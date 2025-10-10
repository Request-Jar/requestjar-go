package store

import (
	"errors"
	"sync"
	"time"

	"github.com/bpietroniro/requestjar-go/internal/models"
	"github.com/bpietroniro/requestjar-go/internal/util"
)

type JarStore interface {
	Create() (string, error)
	Get(id string) (*models.Jar, error)
	List() ([]*models.Jar, error)
}

type jarStore struct {
	jars map[string]*models.Jar
	mu   sync.RWMutex
}

func NewInMemoryJarStore() JarStore {
	return &jarStore{jars: make(map[string]*models.Jar)}
}

func (s *jarStore) Create() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := util.GenerateID()
	s.jars[id] = &models.Jar{ID: id, CreatedAt: time.Now()}

	return id, nil
}

func (s *jarStore) Get(id string) (*models.Jar, error) {
	jar, exists := s.jars[id]
	if !exists {
		return nil, errors.New("jar not found")
	}

	return jar, nil
}

func (s *jarStore) List() ([]*models.Jar, error) {
	jars := make([]*models.Jar, len(s.jars))

	for _, j := range s.jars {
		jars = append(jars, j)
	}

	return jars, nil
}
