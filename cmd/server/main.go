package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/bpietroniro/requestjar-go/internal/logging"
	"github.com/bpietroniro/requestjar-go/internal/router"
	"github.com/bpietroniro/requestjar-go/internal/service"
	"github.com/bpietroniro/requestjar-go/internal/store"
	"github.com/rs/cors"
)

func main() {
	// Logger setup
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logging.LevelTrace,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.LevelKey {
				level := a.Value.Any().(slog.Level)
				levelLabel, exists := logging.LevelNames[level]
				if !exists {
					levelLabel = level.String()
				}

				a.Value = slog.StringValue(levelLabel)
			}

			return a
		},
	}))
	slog.SetDefault(logger)

	// Dependencies
	jarStore := store.NewInMemoryJarStore()
	requestStore := store.NewInMemoryRequestStore()
	svc := service.NewJarService(jarStore, requestStore)
	r := router.CreateRouter(svc)

	// Routing
	mux := http.NewServeMux()

	mux.HandleFunc("GET /jars", r.GetAllJarMetadata)
	mux.HandleFunc("POST /jars", r.CreateJar)
	mux.HandleFunc("DELETE /jars/{jarID}", r.DeleteJar)
	mux.HandleFunc("GET /jars/{jarID}", r.GetJarWithRequests)
	mux.HandleFunc("DELETE /jars/{jarID}/requests/{reqID}", r.DeleteRequest)
	mux.HandleFunc("GET /jars/{jarID}/events", r.HandleSSEConnection)
	mux.HandleFunc("/r/{jarID}/", r.CaptureRequest)
	mux.HandleFunc("/r/{jarID}/{path}", r.CaptureRequest)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprint(w, "hi from Request Jar") // TODO
		if err != nil {
			slog.Error("error writing response", slog.Any("error", err))
		}
	})

	// CORS configuration
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"}, // TODO
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	handler := c.Handler(mux)

	// Server
	slog.Info("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}
