package service

import (
	"log/slog"
	"sync"

	"github.com/bpietroniro/requestjar-go/internal/errors"
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
		return "", err
	}

	err = s.requestStore.CreateJarKey(jarID)
	if err != nil {
		return "", err
	}

	return jarID, nil
}

func (s *JarService) DeleteJar(jarID string) error {
	slog.Info("deleting all requests for jar...", slog.String("jarID", jarID))
	err := s.requestStore.DeleteAllRrequests(jarID)
	if err != nil {
		return err
	}

	slog.Info("deleting jar metadata...", slog.String("jarID", jarID))
	err = s.jarStore.Delete(jarID)
	if err != nil {
		return err
	}

	slog.Info("closing all connections for jar...", slog.String("jarID", jarID))
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
		return nil, nil, err
	}

	requests, err := s.requestStore.List(jarID)
	if err != nil {
		return nil, nil, err
	}

	return jarMetadata, requests, nil
}

func (s *JarService) AddConnection(jarID string, eventChan chan *models.Request) error {
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
	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.connections[jarID]
	if !exists {
		return errors.NotFound("no connections found")
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
	slog.Debug("notifying clients", slog.String("jarID", jarID), slog.String("reqID", request.ID), slog.Int("numConns", len(s.connections[jarID])))
	s.mu.RLock()
	defer s.mu.RUnlock()

	for c := range s.connections[jarID] {
		c <- request
	}
}

func (s *JarService) closeAllConnections(jarID string) {
	slog.Debug("closing all connections", slog.String("jarID", jarID), slog.Int("numConns", len(s.connections[jarID])))
	s.mu.Lock()
	defer s.mu.Unlock()

	if conns, exists := s.connections[jarID]; exists {
		for c := range conns {
			close(c)
		}
		return
	}

	slog.Warn("no connections found, no action taken", slog.String("jarID", jarID))
}
