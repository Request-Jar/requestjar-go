package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/bpietroniro/requestjar-go/internal/router"
	"github.com/bpietroniro/requestjar-go/internal/service"
	"github.com/bpietroniro/requestjar-go/internal/store"
	"github.com/rs/cors"
)

func main() {
	jarStore := store.NewInMemoryJarStore()
	requestStore := store.NewInMemoryRequestStore()
	svc := service.NewJarService(jarStore, requestStore)
	r := router.CreateRouter(svc)

	mux := http.NewServeMux()

	// TODO add handlers
	mux.HandleFunc("GET /jars", r.GetAllJarMetadata)
	mux.HandleFunc("POST /jars", r.CreateJar)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "hi from Request Jar")
	})

	// Configure CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // TODO
		AllowedMethods:   []string{"*"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	handler := c.Handler(mux)

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}
