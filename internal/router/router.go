package router

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/bpietroniro/requestjar-go/internal/errors"
	"github.com/bpietroniro/requestjar-go/internal/models"
	"github.com/bpietroniro/requestjar-go/internal/service"
	"github.com/bpietroniro/requestjar-go/internal/util"
)

type Router struct {
	svc *service.JarService
}

func CreateRouter(svc *service.JarService) *Router {
	slog.Info("creating new router dependency")
	return &Router{svc: svc}
}

func (router *Router) CreateJar(w http.ResponseWriter, r *http.Request) {
	var reqBody CreateJarRequest

	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to parse request body")
		http.Error(w, "error parsing request", http.StatusBadRequest)
		return
	}

	newJarID, err := router.svc.CreateJar(reqBody.Name)
	if err != nil {
		slog.Error("failed to create jar", slog.Any("error", err))
		errors.WriteHTTPError(w, err, "failed to create jar")
		return
	}

	resp := map[string]string{
		"id": newJarID,
	}

	slog.Info("new jar created", slog.String("jarID", newJarID))
	util.WriteJSON(w, http.StatusCreated, resp)
}

func (router *Router) DeleteJar(w http.ResponseWriter, r *http.Request) {
	jarID := r.PathValue("jarID")

	err := router.svc.DeleteJar(jarID)
	if err != nil {
		slog.Error("failed to delete jar", slog.String("jarID", jarID), slog.Any("error", err))
		errors.WriteHTTPError(w, err, "failed to delete jar")
		return
	}

	slog.Info("jar successfully deleted", slog.String("jarID", jarID))
	w.WriteHeader(http.StatusNoContent)
}

func (router *Router) GetAllJarMetadata(w http.ResponseWriter, r *http.Request) {
	jars, err := router.svc.ListAllJarMetadata()

	if err != nil {
		slog.Error("failed to fetch jar metadata", slog.Any("error", err))
		errors.WriteHTTPError(w, err, "failed to fetch jar metadata")
		return
	}

	util.WriteJSON(w, http.StatusOK, jars)
}

func (router *Router) DeleteRequest(w http.ResponseWriter, r *http.Request) {
	jarID := r.PathValue("jarID")
	reqID := r.PathValue("reqID")

	err := router.svc.DeleteRequest(jarID, reqID)

	if err != nil {
		slog.Error("failed to delete jar", slog.String("jarID", jarID), slog.String("reqID", reqID), slog.Any("error", err))
		errors.WriteHTTPError(w, err, "failed to delete request")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (router *Router) GetJarWithRequests(w http.ResponseWriter, r *http.Request) {
	jarID := r.PathValue("jarID")

	jar, requests, err := router.svc.GetJarWithRequests(jarID)

	if err != nil {
		slog.Error("failed to retrieve jar", slog.String("jarID", jarID), slog.Any("error", err))
		errors.WriteHTTPError(w, err, "failed to retrieve jar")
		return
	}

	resp := GetJarWithRequestsResponse{
		Jar:      *jar,
		Requests: requests,
	}

	util.WriteJSON(w, http.StatusCreated, resp)
}

func (router *Router) HandleSSEConnection(w http.ResponseWriter, r *http.Request) {
	jarID := r.PathValue("jarID")

	slog.InfoContext(r.Context(), "adding new SSE connection", slog.String("jarID", jarID))

	// Check that jar exists
	_, err := router.svc.GetJarMetadata(jarID)
	if err != nil {
		slog.Error("jar not found", slog.String("jarID", jarID), slog.Any("error", err))
		errors.WriteHTTPError(w, err, "jar not found")
		return
	}

	// SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		slog.WarnContext(r.Context(), "streaming unsupported, aborting new connection setup")
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	// Create a channel for this requesting client
	eventChan := make(chan *models.Request)

	// Register the connection
	router.svc.AddConnection(jarID, eventChan)

	// Clean up
	defer func() {
		slog.InfoContext(r.Context(), "removing SSE connection", slog.String("jarID", jarID))
		router.svc.RemoveConnection(jarID, eventChan)
		close(eventChan)
	}()

	fmt.Fprintf(w, "data: connected\n\n")
	flusher.Flush()

	// Client has disconnected
	done := r.Context().Done()

	for {
		select {
		case request, ok := <-eventChan:
			// Channel was closed and likely deleted
			if !ok {
				slog.Warn("Channel closed, ending connection")
				return
			}

			// Forward incoming request event to the client
			requestJson, err := json.Marshal(request)
			if err != nil {
				slog.Error("error marshaling request", slog.Any("error", err))
				continue
			}

			slog.Debug("sending request through channel", slog.String("jarID", jarID), slog.String("reqID", request.ID))
			fmt.Fprintf(w, "data: %s\n\n", requestJson)
			flusher.Flush()
		case <-done:
			slog.InfoContext(r.Context(), "Client disconnected", slog.String("jarID", jarID))
			return
		}
	}
}

func (router *Router) CaptureRequest(w http.ResponseWriter, r *http.Request) {
	jarID := r.PathValue("jarID")

	body, err := io.ReadAll(r.Body)

	if err != nil {
		slog.ErrorContext(r.Context(), "failed to read request body")
		http.Error(w, "Failed to read body", http.StatusInternalServerError) // TODO check correct status code
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
		slog.ErrorContext(r.Context(), "failed to create new request", slog.String("jarID", jarID))
		errors.WriteHTTPError(w, err, "failed to create new request")
		return
	}

	slog.Info("request successfully captured", slog.String("jarID", jarID))
}
