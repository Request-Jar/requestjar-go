package store

import "github.com/bpietroniro/requestjar-go/internal/models"

type RequestStore interface {
	Create(jarID string, req *models.Request) error
	Get(jarID string, reqID string) (*models.Request, error)
	List(jarID string) ([]*models.Request, error)
}
