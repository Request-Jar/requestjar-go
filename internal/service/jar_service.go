package service

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/bpietroniro/requestjar-go/internal/models"
	"github.com/bpietroniro/requestjar-go/internal/store"
)

type JarService struct {
	jarStore     store.JarStore
	requestStore store.RequestStore
	connections  map[string]map[chan *models.Request]struct{} // essentially a map of sets
	mu           sync.RWMutex
}

func NewJarService(jarStore store.JarStore, requestStore store.RequestStore) *JarService {
	slog.Info("creating new jar service dependency")
	return &JarService{
		jarStore: jarStore, requestStore: requestStore, connections: make(map[string]map[chan *models.Request]struct{}),
	}
}

func (s *JarService) CreateJar(name string) (string, error) {
	jarID, err := s.jarStore.Create(name)
	if err != nil {
		return "", errors.New("failed to create jar")
	}

	err = s.requestStore.CreateJarKey(jarID)
	if err != nil {
		return "", errors.New("failed to create jar")
	}

	return jarID, nil
}

func (s *JarService) DeleteJar(jarID string) error {
	// TODO handle errors for each step
	s.requestStore.DeleteAllRrequests(jarID)
	s.jarStore.Delete(jarID)
	s.closeAllConnections(jarID)
	return nil
}

func (s *JarService) ListAllJarMetadata() ([]*models.Jar, error) {
	return s.jarStore.List()
}

func (s *JarService) GetJarMetadata(jarID string) (*models.Jar, error) {
	return s.jarStore.Get(jarID)
}

func (s *JarService) GetJarWithRequests(jarID string) (*models.Jar, []*models.Request, error) {
	jarMetadata, err := s.jarStore.Get(jarID)
	if err != nil {
		return nil, nil, errors.Join(err, fmt.Errorf("failed to retrieve jar %s", jarID))
	}

	requests, err := s.requestStore.List(jarID)
	if err != nil {
		return nil, nil, errors.Join(err, fmt.Errorf("failed to retrieve requests for jar %s", jarID))
	}

	return jarMetadata, requests, nil
}

func (s *JarService) AddConnection(jarID string, eventChan chan *models.Request) error {
	slog.Debug("adding connection", slog.String("jarID", jarID))
	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.connections[jarID]
	if !exists {
		s.connections[jarID] = make(map[chan *models.Request]struct{})
	}

	s.connections[jarID][eventChan] = struct{}{}

	return nil
}

func (s *JarService) RemoveConnection(jarID string, eventChan chan *models.Request) error {
	slog.Debug("removing connection", slog.String("jarID", jarID))
	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.connections[jarID]
	if !exists {
		return fmt.Errorf("no connections found for jar %s", jarID)
	}

	delete(s.connections[jarID], eventChan)
	return nil
}

func (s *JarService) NewRequest(jarID string, request *models.Request) error {
	err := s.requestStore.CreateRequest(jarID, request)

	if err != nil {
		return err
	}

	s.notifyClients(jarID, request)

	return nil
}

func (s *JarService) DeleteRequest(jarID string, reqID string) error {
	return s.requestStore.DeleteOneRequest(jarID, reqID)
}

func (s *JarService) notifyClients(jarID string, request *models.Request) {
	slog.Debug("notifying clients", slog.String("jarID", jarID), slog.String("reqID", request.ID))
	s.mu.RLock()
	defer s.mu.RUnlock()

	for c := range s.connections[jarID] {
		c <- request
	}
}

func (s *JarService) closeAllConnections(jarID string) {
	slog.Debug("closing all connections", slog.String("jarID", jarID))
	s.mu.Lock()
	defer s.mu.Unlock()

	if conns, exists := s.connections[jarID]; exists {
		for c := range conns {
			close(c)
		}
	}
}
