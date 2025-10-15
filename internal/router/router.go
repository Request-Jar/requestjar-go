package router

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/bpietroniro/requestjar-go/internal/models"
	"github.com/bpietroniro/requestjar-go/internal/service"
	"github.com/bpietroniro/requestjar-go/internal/util"
)

type Router struct {
	svc *service.JarService
}

func CreateRouter(svc *service.JarService) *Router {
	return &Router{svc: svc}
}

func (router *Router) CreateJar(w http.ResponseWriter, r *http.Request) {
	var reqBody CreateJarRequest

	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		http.Error(w, "error parsing request", http.StatusBadRequest)
		return
	}

	newJarId, err := router.svc.CreateJar(reqBody.Name)
	if err != nil {
		http.Error(w, "failed to create jar", http.StatusInternalServerError)
		return
	}

	resp := map[string]string{
		"id": newJarId,
	}

	util.WriteJSON(w, http.StatusCreated, resp)
}

func (router *Router) DeleteJar(w http.ResponseWriter, r *http.Request) {
	jarID := r.PathValue("jarId")

	err := router.svc.DeleteJar(jarID)
	if err != nil {
		http.Error(w, "failed to delete jar", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (router *Router) GetAllJarMetadata(w http.ResponseWriter, r *http.Request) {
	jars, err := router.svc.ListAllJarMetadata()

	if err != nil {
		http.Error(w, "failed to fetch jar metadata", http.StatusInternalServerError)
		return
	}

	util.WriteJSON(w, http.StatusOK, jars)
}

func (router *Router) DeleteRequest(w http.ResponseWriter, r *http.Request) {
	// jarID := r.PathValue("jarID")
	// reqID := r.PathValue("reqID")
	// TODO
}

func (router *Router) GetJarWithRequests(w http.ResponseWriter, r *http.Request) {
	// TODO
}

func (router *Router) HandleSSEConnection(w http.ResponseWriter, r *http.Request) {
	jarID := r.PathValue("jarID")

	// Check that jar exists
	_, err := router.svc.GetJarMetadata(jarID)
	if err != nil {
		http.Error(w, "Jar not found", http.StatusNotFound)
		return
	}

	// SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Create a channel for this requesting client
	eventChan := make(chan *models.Request)

	// Register the connection
	router.svc.AddConnection(jarID, eventChan)

	// Clean up
	defer func() {
		router.svc.RemoveConnection(jarID, eventChan)
		close(eventChan)
	}()

	// Client has disconnected
	done := r.Context().Done()

	for {
		select {
		case request := <-eventChan:
			// Forward incoming request event to the client
			json.NewEncoder(w).Encode(request)
		case <-done:
			return
		}
	}
}

func (router *Router) CaptureRequest(w http.ResponseWriter, r *http.Request) {
	jarID := r.PathValue("jarId")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	headers := make(map[string]string)
	for key, values := range r.Header {
		headers[key] = values[0] // TODO verify this
	}

	query := make(map[string]string)
	for key, values := range r.URL.Query() {
		query[key] = values[0] // TODO verify this
	}

	req := &models.Request{
		CreatedAt: time.Now(),
		Method:    r.Method,
		Path:      r.URL.Path,
		Headers:   headers,
		Query:     query,
		Body:      body,
		ClientIP:  r.RemoteAddr,
	}

	err = router.svc.NewRequest(jarID, req)

	if err != nil {
		// TODO
	}
}
