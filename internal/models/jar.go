package models

import "time"

type Jar struct {
	ID        string
	CreatedAt time.Time
	Requests  []Request
}

type Request struct {
	ID        string
	CreatedAt time.Time
	Method    string
	Path      string
	Headers   map[string]string
	ClientIP  string
	Body      []byte
	Query     map[string]string
}
