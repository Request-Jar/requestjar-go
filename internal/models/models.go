package models

import "time"

type Jar struct {
	ID        string
	Name      string
	CreatedAt time.Time
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
