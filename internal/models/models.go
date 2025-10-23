package models

import "time"

type Jar struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}

type Request struct {
	ID        string            `json:"id"`
	CreatedAt time.Time         `json:"createdAt"`
	Method    string            `json:"method"`
	Path      string            `json:"path"`
	Headers   map[string]string `json:"headers"`
	ClientIP  string            `json:"clientIP"`
	Body      []byte            `json:"body"`
	Query     map[string]string `json:"query"`
}
