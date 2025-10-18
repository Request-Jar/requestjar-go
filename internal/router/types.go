package router

import "github.com/bpietroniro/requestjar-go/internal/models"

type CreateJarRequest struct {
	Name string `json:"name"`
}

type DeleteJarRequest struct {
	ID string `json:"id"`
}

type DeleteRequestRequest struct {
	ID string `json:"id"`
}

type GetJarWithRequestsResponse struct {
	Jar      models.Jar        `json:"jar"`
	Requests []*models.Request `json:"requests"` // TODO pointers or not?
}
